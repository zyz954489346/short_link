package libraries

import (
	"context"
	"errors"
	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var rdb = make(map[int]*redis.Client)

// getDefaultDB 获取默认的 db 编号
func getDefaultDB(dbIndex *int) int {
	if dbIndex == nil {
		di, _ := strconv.ParseInt(
			os.Getenv("REDIS_DATABASE_DEFAULT"),
			10,
			8,
		)
		return int(di)
	}

	return *dbIndex
}

// RedisConn 连接 redis 对应 db
func RedisConn(dbIndex *int) {
	di := getDefaultDB(dbIndex)

	addr := strings.Join([]string{
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
	}, ":")

	// 启动连接池
	rdb[di] = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           di,
		PoolSize:     10,               // 最大连接数
		MinIdleConns: 5,                // 最小空闲连接数
		DialTimeout:  10 * time.Second, // 连接超时
		ReadTimeout:  10 * time.Second, // 读超时
		WriteTimeout: 10 * time.Second, // 写超时
	})
}

// RedisDisConn 关闭所有 redis 连接
func RedisDisConn() {
	if rdb == nil {
		return
	}

	for k, _ := range rdb {
		if err := rdb[k].Close(); err != nil {
			panic("关闭 Redis 连接失败：" + err.Error())
		}
	}
}

// Redis 获取 redis 对应 db 的连接
func Redis(dbIndex *int) *redis.Client {
	di := getDefaultDB(dbIndex)

	if rdb[di] == nil {
		RedisConn(&di)
	}

	return rdb[di]
}

// withRedisLock 在 Redis 分布式锁周期下进行相关操作
func withRedisLock(c context.Context, key string, handler func() (any, error), ttl *time.Duration) (any, error) {
	if ttl == nil {
		defaultTtl := time.Millisecond * 500
		ttl = &defaultTtl
	}

	// 锁创建
	locker := redislock.New(Redis(nil))

	// 锁获取
	lock, err := locker.Obtain(c, Md5(key), *ttl, nil)
	// 锁释放
	defer func(lock *redislock.Lock, ctx context.Context) {
		_ = lock.Release(ctx)
	}(lock, c)

	if errors.Is(err, redislock.ErrNotObtained) {
		return "", errors.New("获取 redis 锁失败")
	} else if err != nil {
		return "", errors.New("获取 redis 锁失败：" + err.Error())
	}

	return handler()
}

// buildRedisKey 构建 redis 缓存 key
func buildRedisKey(key string) string {
	return strings.Join([]string{
		"Gin",
		os.Getenv("APP_NAME"),
		os.Getenv("APP_ENV"),
		Md5(url.QueryEscape(key)),
	}, ":")
}

// CacheSet 缓存存操作
func CacheSet(ctx context.Context, key string, val string, ttl time.Duration) (string, error) {
	return Redis(nil).Set(ctx, buildRedisKey(key), val, ttl).Result()
}

// CacheGet 缓存取操作
func CacheGet(ctx context.Context, key string) (string, error) {
	return Redis(nil).Get(ctx, buildRedisKey(key)).Result()
}

// CacheSetWithLock 带锁的缓存存
func CacheSetWithLock(ctx context.Context, key string, val string, ttl time.Duration) (string, error) {
	cacheRes, err := withRedisLock(ctx, key, func() (any, error) {
		return CacheSet(ctx, key, val, ttl)
	}, nil)

	if err != nil {
		return "", err
	}
	return cacheRes.(string), nil
}

// CacheGetWithLock 带锁的缓存读
func CacheGetWithLock(ctx context.Context, key string) (string, error) {
	cacheRes, err := withRedisLock(ctx, key, func() (any, error) {
		return CacheGet(ctx, key)
	}, nil)

	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return cacheRes.(string), nil
}

// RememberWithLock 带锁的读，在不存在的时候自动处理写操作
func RememberWithLock(ctx context.Context, key string, valFnc func() string, ttl time.Duration) (interface{}, error) {
	return withRedisLock(ctx, key, func() (any, error) {
		// 查缓存
		cacheRes, getErr := CacheGet(ctx, key)
		if getErr != nil {
			if getErr == redis.Nil {
				return "", nil
			}
			return "", getErr
		}

		// 存在
		if cacheRes == "" {
			// 不存在
			var setErr error
			cacheRes, setErr = CacheSet(ctx, key, valFnc(), ttl)

			if setErr != nil {
				return "", setErr
			}
		}

		return cacheRes, nil
	}, nil)
}
