package endpoints

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"short_link/business"
	"short_link/libraries"
	"strings"
)

// LongUrl 长地址
type LongUrl struct {
	Url  string `json:"url" binding:"required"`
	Key  string `json:"key"`
	Sign string `json:"sign"`
}

// ShortenUrl 使长地址转短地址
func ShortenUrl(c *gin.Context) {
	var longUrl LongUrl
	if err := c.ShouldBindBodyWithJSON(&longUrl); err != nil {
		libraries.Err(c, &libraries.Response{Message: err.Error()})
		return
	}

	// 缩短后的连接
	shortUrl, urlErr := business.MakeUrlShorter(c, longUrl.Url)

	if urlErr != nil {
		libraries.Err(c, &libraries.Response{Message: urlErr.Error()})
		return
	}

	shortUrl = strings.Join([]string{
		os.Getenv("APP_URL"),
		shortUrl,
	}, "/urls/v/")

	libraries.Ok(c, &libraries.Response{Data: shortUrl})
}

// VisitShortUrl 访问短地址
func VisitShortUrl(c *gin.Context) {

	shortUrl := c.Param("shortUrl")
	var visitUrl string
	var err error

	if visitUrl, err = business.GetVisitUrl(c, shortUrl); err != nil {
		visitUrl = shortUrl
	}

	c.Redirect(http.StatusFound, visitUrl)
}
