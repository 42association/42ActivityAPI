package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"net/http"
	"os"
)


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
