package routes

import (
	"github.com/gin-gonic/gin"
	"short_link/endpoints"
	"short_link/middlewares"
)

func Register(r *gin.Engine) {
	// 路由分组
	router := r.Group("/urls")
	{
		// 传入长路由，输出短路由
		router.POST("/shorten", middlewares.SignCheck, endpoints.ShortenUrl)

		// 传入短路由，直接重定向
		router.GET("/v/:shortUrl", endpoints.VisitShortUrl)
	}
}
