package db

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
)

var RedisClient *redis.Client

func init() {
	REDIS_ADDR := os.Getenv("REDIS_ADDR")
	REDIS_PASSWORD := os.Getenv("REDIS_PASSWORD")

	if REDIS_ADDR == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}

	if REDIS_PASSWORD == "" {
		log.Fatal("REDIS_PASSWORD environment variable is not set")
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PASSWORD,
		Username: "default",
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
}
