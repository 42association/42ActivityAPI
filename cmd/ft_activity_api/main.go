package main

import (
	"42ActivityAPI/internal/accessdb"
	"42ActivityAPI/internal/handlers"
	"42ActivityAPI/internal/loadconfig"
	"42ActivityAPI/internal/logging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"log/slog"
	"net/http"
	"os"
	"io"
)

func main() {
	var file *os.File
	var err error
	file, err = os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open log file: ", err)
		return
	}
	defer file.Close()
	writter := io.Writer(file)
	logging.Logger = slog.New(slog.NewJSONHandler(writter, nil))
	slog.SetDefault(logging.Logger)

	// Initialize database
	_, err = accessdb.ConnectToDB()
	if err != nil {
		log.Println("Failed to initialize database: ", err)
		return
	}

	router := gin.Default()
	router.Use(logging.GinLogger(logging.Logger))

	router.LoadHTMLGlob("web/templates/*")

	// CORS Settings
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	router.GET("/", ShowIndexPage)
	router.GET("/new", RedirectToIndexWithUID)
	router.GET("/callback", ShowCallbackPage)

	router.POST("/receive-uid", handlers.HandleUIDSubmission)

	router.GET("/shifts", handlers.GetShiftData)
	router.POST("/shifts", handlers.AddShiftData)
	router.POST("/shifts/exchange", handlers.ExchangeShiftData)
	router.DELETE("/shifts", handlers.DeleteShiftData)

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
