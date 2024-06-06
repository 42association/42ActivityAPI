package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/42association/42ActivityAPI/internal/handlers/handler"
	"github.com/42association/42ActivityAPI/internal/accessdb/accessdb"
)

type Config struct {
	UID         string
	Secret      string
	CallbackURL string
}

func main() {
	// Initialize database
	_, err := connectToDB();
	if err != nil {
		log.Println("Failed to initialize database: ", err)
		return
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// CORS Settings
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	router.GET("/", ShowIndexPage)
	router.GET("/new", RedirectToIndexWithUID)
	router.GET("/callback", ShowCallbackPage)
	router.POST("/receive-uid", HandleUIDSubmission)

	router.GET("/shift", getShiftData)
	router.POST("/shift", addShiftData)

	router.POST("/activities", addActivity)
	router.GET("/activities/cleanings", getActivityCleanData)

	router.POST("/roles", addRole)

	router.POST("/locations", addLocation)

	router.POST("/m5sticks", addM5Stick)

	router.POST("/users", addUser)
	router.PUT("/users", editUser)

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
