package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/araviindm/url-shortener/db"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShortenURLRequest struct {
	LongURL string `json:"long_url"`
}

type URLLongShortMapping struct {
	LongURL  string `bson:"long_url" json:"long_url"`
	ShortURL string `bson:"short_url" json:"short_url"`
}

func ShortenURL(c *gin.Context) {
	var req ShortenURLRequest

	// Parsing JSON request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request"})
		return
	}
	// Checking if the long URL exists in Redis cache or MongoDB
	shortURL, err := checkURLExistence(req.LongURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check URL existence"})
		return
	}

	if shortURL != "" {
		c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
		return
	}

	// Long URL does not exist, generate short URL
	shortURL = generateShortURL(req.LongURL)

	// Cache the mapping in Redis
	err = db.RedisClient.Set(context.Background(), req.LongURL, shortURL, 0).Err()
	if err != nil {
		log.Println("Failed to store mapping in Redis:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	// Store the mapping in MongoDB
	collection := db.MongoClient.Collection("url_mappings")
	_, err = collection.InsertOne(context.Background(), bson.M{"short_url": shortURL, "long_url": req.LongURL})
	if err != nil {
		log.Println("Failed to store mapping in MongoDB:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url": shortURL,
	})
}

func generateShortURL(longURL string) string {
	// Calculate MD5 hash of the long URL
	hash := md5.Sum([]byte(longURL))

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hash[:])

	// Take the first 8 characters of the hash string as the short URL
	shortURL := hashString[:10]
	log.Println("Created")
	return shortURL
}

func checkURLExistence(longURL string) (string, error) {

	ctx := context.Background()
	// Check if the long URL already exists in Redis cache
	shortURL, err := db.RedisClient.Get(ctx, longURL).Result()
	if err == nil {
		log.Println("In Redis")
		return shortURL, nil
	} else if err != redis.Nil {
		log.Println("Redis error:", err)
		return "", err
	}

	//  Check if the long URL already exists in MongoDB
	collection := db.MongoClient.Collection("url_mappings")
	var mapping URLLongShortMapping
	err = collection.FindOne(ctx, bson.M{"long_url": longURL}).Decode(&mapping)
	if err == nil {
		log.Println("In Mongo")
		return mapping.ShortURL, nil
	} else if err != mongo.ErrNoDocuments {
		log.Println("MongoDB error:", err)
		return "", err
	}
	return "", nil
}

func RedirectToOriginalURL(c *gin.Context) {
}
