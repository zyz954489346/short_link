package business

import (
	"errors"
	"github.com/gin-gonic/gin"
	"short_link/libraries"
	"short_link/models"
	"strconv"
	"sync"
	"time"
)

// 带锁的 url 操作
type urlStoreLocker struct {
	sync.RWMutex // 读写锁
	url          string
}

// saveUrlRelationToCache 缓存长短URL关系至缓存
func saveUrlRelationToCache(c *gin.Context, longUrl string, shortUrl string, ttl time.Duration) (string, error) {
	_, slErr := libraries.CacheSetWithLock(c.Request.Context(), shortUrl, longUrl, ttl)
	if slErr != nil {
		return "", slErr
	}
	_, lsErr := libraries.CacheSetWithLock(c.Request.Context(), longUrl, shortUrl, ttl)
	if lsErr != nil {
		return "", lsErr
	}

	return shortUrl, nil
}

// MakeUrlShorter 长链接转短连接
// 先取redis，再取mysql，这些步骤需要加锁
func MakeUrlShorter(c *gin.Context, url string) (string, error) {
	// 协程写锁对象
	store := urlStoreLocker{url: url}

	// 协程读取锁
	var shortUrl string
	var getErr error
	if shortUrl, getErr = func() (string, error) {

		// 协程读锁
		store.RLock()
		defer store.RUnlock()
		// 读缓存
		return libraries.CacheGetWithLock(c.Request.Context(), url)
	}(); getErr != nil {
		return "", getErr
	}

	// 缓存没有就查库
	if shortUrl == "" {
		var setErr error
		if shortUrl, setErr = func() (string, error) {
			store.Lock()
			defer store.Unlock()
			models.
				Database().
				Model(&models.UrlStore{}).
				Select("short_url").
				Where(map[string]string{"url": url}).
				Scan(&shortUrl)

			// 查库的结果写入缓存
			if shortUrl != "" {
				return saveUrlRelationToCache(c, url, shortUrl, 0)
			}

			return shortUrl, nil
		}(); setErr != nil {
			return "", setErr
		}
	}

	// 缓存和数据库都查不到，则生成一个
	if shortUrl == "" {
		// 协程写锁
		store.Lock()
		defer store.Unlock()
		// 12 位唯一编号
		shortUrl = libraries.Nanoid(12)
		// 存库
		appId, _ := strconv.ParseUint(c.Request.Header.Get("X-APP-ID"), 10, 8)
		urlStore := models.UrlStore{
			AppId:    uint(appId),
			Url:      url,
			ShortUrl: shortUrl,
		}

		insertCount := models.Database().Create(&urlStore)

		if insertCount.Error != nil {
			return "", errors.New("生成短链失败")
		}
		// 存 redis
		return saveUrlRelationToCache(c, url, shortUrl, 0)
	}

	return shortUrl, nil
}

// GetVisitUrl 通过短地址获取实际的跳转地址
func GetVisitUrl(c *gin.Context, shortUrl string) (string, error) {

	var url string
	store := urlStoreLocker{url: shortUrl}

	// 使用短地址查长地址
	var getErr error
	if url, getErr = func() (string, error) {
		// 协程读锁
		store.RLock()
		defer store.RUnlock()

		// 查 redis
		return libraries.CacheGetWithLock(c.Request.Context(), shortUrl)
	}(); getErr != nil {
		return "", getErr
	}

	if url == "" {
		var setErr error
		if url, setErr = func() (string, error) {
			// 协程写锁
			store.Lock()
			defer store.Unlock()
			// redis 不存在就查 mysql
			models.Database().Select("url").Where(map[string]string{"short_url": shortUrl}).Scan(&url)

			if url == "" {
				return "", errors.New("地址不存在")
			}

			// mysql 的结果写入 redis
			_, cacheErr := libraries.CacheSetWithLock(c.Request.Context(), shortUrl, url, 0)
			if cacheErr != nil {
				return "", cacheErr
			}

			return url, nil
		}(); setErr != nil {
			return "", setErr
		}
	}

	return url, nil
}
