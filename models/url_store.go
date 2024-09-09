package models

import "gorm.io/gorm"

// UrlStore 短链配置
type UrlStore struct {
	gorm.Model
	AppId    uint   `gorm:"type:bigint;not null;default:0"`
	Url      string `gorm:"index:unique;size:255;default:''"`
	ShortUrl string `gorm:"index:unique;size:100;default:''"`
}

// TableName 指定表名
func (*UrlStore) TableName() string {
	return "url_stores"
}

// GetShortUrlByUrl 长链接查短链接
func (urlStore *UrlStore) GetShortUrlByUrl(url string) (shortUrl string) {
	where := map[string]interface{}{
		"url": url,
	}
	Database().Model(urlStore).Select("short_url").Where(where).Scan(&shortUrl)

	return shortUrl
}

// GetUrlByShortUrl 短链接查长链接
func (urlStore *UrlStore) GetUrlByShortUrl(shortUrl string) (url string) {
	where := map[string]interface{}{
		"short_url": shortUrl,
	}
	Database().Model(urlStore).Select("url").Where(where).Scan(&url)

	return url
}
