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
	mutex sync.Mutex // Mutex to synchronize access to shared resources
)

type ShortenURLRequest struct {
	LongURL string `json:"long_url"`
}

type URLLongShortMapping struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	LongURL  string             `bson:"long_url" json:"long_url"`
	ShortURL string             `bson:"short_url" json:"short_url"`
}

// Result represents the result of get from redis, mongdb function
type Result struct {
	LongURL string
	Err     error
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

	// Create channels for communication
	redisResult := make(chan error)
	mongoResult := make(chan Result)
	var redisErr error
	var mongoResp Result

	go func() {
		// Cache the mapping in Redis
		err := db.RedisClient.Set(context.Background(), shortURL, req.LongURL, 0).Err()
		redisResult <- err
	}()

	go func() {
		// Check if already exists
		resp := checkMongoForURL(shortURL)
		mongoResult <- resp
	}()

	redisErr = <-redisResult
	mongoResp = <-mongoResult

	if redisErr != nil {
		log.Println("Failed to store mapping in Redis:", redisErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	if mongoResp.Err != nil {
		log.Println("Error checking MongoDB for URL", mongoResp.Err)
	}
	// Construct the full URL and return it
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http" // Default to HTTP if X-Forwarded-Proto header is not set
	}
	baseURL := scheme + "://" + c.Request.Host
	fullURL := fmt.Sprintf("%s/%s", baseURL, shortURL)

	if mongoResp.LongURL != "" {
		log.Println("In Mongo")
		c.JSON(http.StatusOK, gin.H{
			"short_url": fullURL,
			"long_url":  req.LongURL,
		})
		return
	}

	insertResult := make(chan error)
	var insertErr error

	go func() {
		// Store the mapping in MongoDB
		collection := db.MongoClient.Collection("url_mappings")
		_, err := collection.InsertOne(context.Background(), bson.M{"short_url": shortURL, "long_url": req.LongURL})
		insertResult <- err
	}()
	insertErr = <-insertResult
	if insertErr != nil {
		log.Println("Failed to store mapping in MongoDB:", insertErr)
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

	redisResult := make(chan Result)
	mongoResult := make(chan Result)

	var redisResp Result
	var mongoResp Result
	go func() {
		// Check Redis cache for the short URL mapping
		longURL, err := db.RedisClient.Get(context.Background(), shortURL).Result()
		redisResult <- Result{longURL, err}
	}()
	go func() {
		resp := checkMongoForURL(shortURL)
		mongoResult <- resp
	}()

	redisResp = <-redisResult
	mongoResp = <-mongoResult

	if redisResp.Err == nil {
		log.Println("In Redis")
		c.Redirect(http.StatusFound, redisResp.LongURL)
		return
	} else if redisResp.Err != redis.Nil {
		log.Println("Redis error:", redisResp.Err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original URL"})
		return
	}

	if mongoResp.Err != nil {
		log.Println("Error checking MongoDB for URL:", mongoResp)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve original URL"})
		return
	}

	if mongoResp.LongURL != "" {
		log.Println("In Mongo")
		c.Redirect(http.StatusFound, mongoResp.LongURL)
		return
	}

	// Short URL mapping not found in Redis, check MongoDB
	c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
}

func checkMongoForURL(shortURL string) Result {
	collection := db.MongoClient.Collection("url_mappings")
	var mapping URLLongShortMapping
	err := collection.FindOne(context.Background(), bson.M{"short_url": shortURL}).Decode(&mapping)
	if err == nil {
		log.Println("Short URL exists in MongoDB")
		return Result{mapping.LongURL, nil}
	} else if err == mongo.ErrNoDocuments {
		log.Println("Short URL does not exist in MongoDB")
		return Result{"", err}
	} else {
		log.Println("Error while checking MongoDB for URL:", err)
		return Result{"", err}
	}
}
