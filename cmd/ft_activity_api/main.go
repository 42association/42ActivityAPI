package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"42ActivityAPI/internal/handlers"
	"42ActivityAPI/internal/accessdb"
)

type Config struct {
	UID         string
	Secret      string
	CallbackURL string
}

func main() {
	// Initialize database
	_, err := ConnectToDB();
	if err != nil {
		log.Println("Failed to initialize database: ", err)
		return
	}

	router := gin.Default()
	router.LoadHTMLGlob("../../web/templates/*")

	// CORS Settings
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	router.GET("/", ShowIndexPage)
	router.GET("/new", RedirectToIndexWithUID)
	router.GET("/callback", ShowCallbackPage)
	router.POST("/receive-uid", HandleUIDSubmission)

	router.GET("/shift", GetShiftData)
	router.POST("/shift", AddShiftData)

	router.POST("/activities", AddActivity)
	router.GET("/activities/cleanings", GetActivityCleanData)

	router.POST("/roles", AddRole)

	router.POST("/locations", AddLocation)

	router.POST("/m5sticks", AddM5Stick)

	router.POST("/users", AddUser)
	router.PUT("/users", EditUser)

	router.Run(":" + os.Getenv("PORT"))
}

// Loading environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		UID:         os.Getenv("UID"),
		Secret:      os.Getenv("SECRET"),
		CallbackURL: os.Getenv("CALLBACK_URL"),
	}
	if config.UID == "" || config.Secret == "" || config.CallbackURL == "" {
		return nil, errors.New("one or more required environment variables are not set")
	}
	return config, nil
}

func ShowIndexPage(c *gin.Context) {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"UID":          config.UID,
		"CALLBACK_URL": config.CallbackURL,
	})
}

func RedirectToIndexWithUID(c *gin.Context) {
	uid := c.Query("uid")
	c.Redirect(http.StatusMovedPermanently, "/?uid="+uid)
}

func ShowCallbackPage(c *gin.Context) {
	c.HTML(http.StatusOK, "callback.html", nil)
}
