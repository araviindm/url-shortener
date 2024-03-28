package main

import (
	"github.com/araviindm/url-shortener/api"
	"github.com/gin-gonic/gin"
)

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.POST("api/shorten", api.ShortenURL)
	router.GET("/*shortURL", api.RedirectToOriginalURL)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
