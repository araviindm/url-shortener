package main

import (
	"github.com/araviindm/url-shortener/api"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Define routes
	router.POST("/shorten", api.ShortenURL)
	router.GET("/:shortURL", api.RedirectToOriginalURL)

	// Run the server
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
