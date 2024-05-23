package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"net/http"
	"os"
)

func TestMain(m *testing.M) {
	db, _ := initializeDB();
	seed(db)
	code := m.Run()
	os.Exit(code)
}

func TestGetQueryAboutTime(t *testing.T) {
    req, _ := http.NewRequest("GET", "/activities/cleanings?start=100&end=200", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	start, end, _ := getQueryAboutTime(c)
	assert.Equal(t, int64(100), start)
	assert.Equal(t, int64(200), end)
}


func TestLoadConfig(t *testing.T) {
	config, _ := LoadConfig()
	assert.Equal(t, os.Getenv("UID"), config.UID)
	assert.Equal(t, os.Getenv("SECRET"), config.Secret)
	assert.Equal(t, os.Getenv("CALLBACK_URL"), config.CallbackURL)
}

type MockConfig struct {
    UID         string
    CallbackURL string
}

// Mock LoadConfig function for testing purposes
func MockLoadConfig() (*MockConfig, error) {
    // Mocked configuration values
    config := &MockConfig{
        UID:         "mockUID",
        CallbackURL: "http://mock-callback-url.com",
    }
    return config, nil
}

func TestShowIndexPage(t *testing.T) {
    router := gin.New()
    router.LoadHTMLGlob("templates/*")

    router.GET("/", ShowIndexPage)

    req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatal(err)
    }

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

}

func TestRedirectToIndexWithUID(t *testing.T) {
	router := gin.New()

	router.GET("/redirect/new", RedirectToIndexWithUID)

	req, err := http.NewRequest("GET", "/redirect/new?uid=123", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	expectedRedirectURL := "/?uid=123"
	assert.Equal(t, expectedRedirectURL, w.Header().Get("Location"))
}

func TestShowCallbackPage(t *testing.T) {
	router := gin.New()

	router.LoadHTMLGlob("templates/*.html")
	router.GET("/callback", ShowCallbackPage)

	req, err := http.NewRequest("GET", "/callback", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestGetShiftUsers(t *testing.T) {
	shifts, _ := getShiftFromDB("2024-06-01")
	assert.Equal(t, "kakiba", shifts[0].User.Login)
}

func TestAddDuplicated(t *testing.T) {
	assert.Equal(t, true, userExists("kakiba"))
	assert.Equal(t, false, userExists("anonymous"))
}
