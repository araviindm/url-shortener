package main

import (
	"log"
	"os"

	"github.com/araviindm/url-shortener/api"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.POST("api/shorten", api.ShortenURL)
	router.GET("/*shortURL", api.RedirectToOriginalURL)

	PORT := ":" + os.Getenv("PORT")

	if PORT == "" {
		log.Fatal("PORT environment variable is not set")
	}

	if err := router.Run(PORT); err != nil {
		panic(err)
	}

}
