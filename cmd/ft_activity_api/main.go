package main

import (
	"42ActivityAPI/internal/accessdb"
	"42ActivityAPI/internal/handlers"
	"42ActivityAPI/internal/loadconfig"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func main() {
	// Initialize database
	_, err := accessdb.ConnectToDB()
	if err != nil {
		log.Println("Failed to initialize database: ", err)
		return
	}

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")

	// CORS Settings
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	router.GET("/", ShowIndexPage)
	router.GET("/new", RedirectToIndexWithUID)
	router.GET("/callback", ShowCallbackPage)

	router.POST("/receive-uid", handlers.HandleUIDSubmission)

	router.GET("/shift", handlers.GetShiftData)
	router.POST("/shift", handlers.AddShiftData)
	router.POST("/shift/exchange", handlers.ExchangeShiftData)

	router.POST("/activities", handlers.AddActivity)
	router.GET("/activities/cleanings", handlers.GetActivityCleanData)

	router.POST("/roles", handlers.AddRole)

	router.POST("/locations", handlers.AddLocation)

	router.POST("/m5sticks", handlers.AddM5Stick)

	router.POST("/users", handlers.AddUsers)
	router.PUT("/users", handlers.EditUser)

	router.Run(":" + os.Getenv("PORT"))
}

func ShowIndexPage(c *gin.Context) {
	config, err := loadconfig.LoadConfig()
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
