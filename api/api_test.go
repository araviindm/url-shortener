package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Unit test for generateShortURL function
func TestGenerateShortURL(t *testing.T) {
	longURL := "https://www.example.com"

	shortURL := GenerateShortURL(longURL)

	// Assert that the generated short URL is not empty
	assert.NotEmpty(t, shortURL, "Generated short URL should not be empty")

	// Assert that the length of the generated short URL is as expected
	assert.Equal(t, 10, len(shortURL), "Generated short URL length should be 10")
}

func TestShortenURL(t *testing.T) {
	// Initialize a new Gin router
	router := gin.Default()

	// Define a sample payload
	payload := []byte(`{"long_url": "https://example.com"}`)

	// Create a POST request with the sample payload
	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Set Content-Type header to JSON
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Set up the handler function for testing
	router.POST("/api/shorten", ShortenURL)

	// Perform the request
	router.ServeHTTP(rr, req)

	// Check if the status code is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	shortURL := GenerateShortURL("https://example.com")

	// Define the expected response as a map[string]interface{}
	expectedResponse := map[string]interface{}{
		"short_url": "http://localhost:8080/" + shortURL,
		"long_url":  "https://example.com",
	}

	// Parse the actual response body into a map[string]interface{}
	var actualResponse map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedResponse, actualResponse)
}

func TestRedirectToOriginalURL(t *testing.T) {
	// Initialize a new Gin router
	router := gin.Default()

	// Register the RedirectToOriginalURL handler
	router.GET("/*shortURL", RedirectToOriginalURL)

	// Define a sample short URL and its corresponding long URL
	shortURL := "100680ad54"
	longURL := "https://example.com"

	// Create a request with the sample short URL
	req, err := http.NewRequest("GET", "/"+shortURL, nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(rr, req)

	// Check if the status code is 302 Found (redirect)
	if rr.Code != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, rr.Code)
	}

	// Check if the location header matches the expected long URL
	expectedLocation := longURL
	if rr.Header().Get("Location") != expectedLocation {
		t.Errorf("Expected Location header %q, got %q", expectedLocation, rr.Header().Get("Location"))
	}
}
