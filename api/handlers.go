package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/araviindm/url-shortener/db"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mutex       sync.Mutex          // Mutex to synchronize access to shared resources
	redirectURL = make(chan string) // Channel for communicating redirection URLs
)

type ShortenURLRequest struct {
	LongURL string `json:"long_url"`
}

type URLLongShortMapping struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	LongURL  string             `bson:"long_url" json:"long_url"`
	ShortURL string             `bson:"short_url" json:"short_url"`
}

func ShortenURL(c *gin.Context) {
	var req ShortenURLRequest

	// Parsing JSON request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request"})
		return
	}

	// Long URL does not exist, generate short URL
	shortURL := generateShortURL(req.LongURL)

	mutex.Lock()         // Lock before accessing shared resources
	defer mutex.Unlock() // Ensure mutex is released

	// Cache the mapping in Redis
	err := db.RedisClient.Set(context.Background(), shortURL, req.LongURL, 0).Err()
	if err != nil {
		log.Println("Failed to store mapping in Redis:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	// Check if already exists
	existingLongURL, err := checkMongoForURL(shortURL)
	if err != nil {
		log.Println("Error checking MongoDB for URL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original URL"})
		return
	}
	// Construct the full URL and return it
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http" // Default to HTTP if X-Forwarded-Proto header is not set
	}
	baseURL := scheme + "://" + c.Request.Host
	fullURL := fmt.Sprintf("%s/%s", baseURL, shortURL)

	if existingLongURL != "" {
		log.Println("In Mongo")
		c.JSON(http.StatusOK, gin.H{
			"short_url": fullURL,
			"long_url":  req.LongURL,
		})
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
		"short_url": fullURL,
		"long_url":  req.LongURL,
	})
}

func generateShortURL(longURL string) string {
	// Calculate SHA-256 hash of the long URL
	hash := sha256.Sum256([]byte(longURL))

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hash[:])

	// Take the first 10 characters of the hash string as the short URL
	shortURL := hashString[:10]
	log.Println("Created")
	return shortURL
}

func RedirectToOriginalURL(c *gin.Context) {
	shortURL := strings.TrimPrefix(c.Param("shortURL"), "/")

	mutex.Lock()         // Lock before accessing shared resources
	defer mutex.Unlock() // Ensure mutex is released

	// Check Redis cache for the short URL mapping
	longURL, err := db.RedisClient.Get(context.Background(), shortURL).Result()
	if err == nil {
		log.Println("In Redis")

		go processRedirect(c, longURL)
		return
	} else if err != redis.Nil {
		log.Println("Redis error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original URL"})
		return
	}

	existingLongURL, err := checkMongoForURL(shortURL)
	if err != nil {
		log.Println("Error checking MongoDB for URL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original URL"})
		return
	}

	if existingLongURL != "" {
		log.Println("In Mongo")
		c.Redirect(http.StatusFound, existingLongURL)
		return
	}

	// Short URL mapping not found in Redis, check MongoDB

	c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
}

func checkMongoForURL(shortURL string) (string, error) {
	collection := db.MongoClient.Collection("url_mappings")
	var mapping URLLongShortMapping
	err := collection.FindOne(context.Background(), bson.M{"short_url": shortURL}).Decode(&mapping)
	if err == nil {
		log.Println("Short URL exists in MongoDB")
		return mapping.LongURL, nil
	} else if err == mongo.ErrNoDocuments {
		log.Println("Short URL does not exist in MongoDB")
		return "", nil
	} else {
		log.Println("Error while checking MongoDB for URL:", err)
		return "", err
	}
}

func processRedirect(c *gin.Context, longURL string) {
	url := <-redirectURL
	c.Redirect(http.StatusFound, url)
}
