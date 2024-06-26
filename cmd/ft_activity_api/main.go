package main

import (
	"42ActivityAPI/internal/accessdb"
	"42ActivityAPI/internal/checkauth"
	"42ActivityAPI/internal/handlers"
	"42ActivityAPI/internal/loadconfig"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"encoding/json"
	"strings"
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
	// router.GET("/callback", ShowCallbackPage)
	router.GET("/callback", handleCallback)

	authGroup := router.Group("/", checkauth.CheckAuthHeader())
	{
		authGroup.POST("/receive-uid", handlers.HandleUIDSubmission)

		authGroup.GET("/shifts", handlers.GetShiftData)
		authGroup.POST("/shifts", handlers.AddShiftData)
		authGroup.POST("/shifts/exchange", handlers.ExchangeShiftData)
		authGroup.DELETE("/shifts", handlers.DeleteShiftData)

		authGroup.POST("/activities", handlers.AddActivity)
		authGroup.GET("/activities/cleanings", handlers.GetActivityCleanData)

		authGroup.POST("/roles", handlers.AddRole)

		authGroup.POST("/locations", handlers.AddLocation)

		authGroup.POST("/m5sticks", handlers.AddM5Stick)

		authGroup.POST("/users", handlers.AddUsers)
		authGroup.PUT("/users", handlers.EditUser)
	}

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

func handleCallback(c *gin.Context) {
	uid := c.Query("uid")
	code := c.Query("code")
	if uid == "" || code == "" {
		c.HTML(http.StatusBadRequest, "response.html", gin.H{"Message": "NFCタグとログインの紐付けに失敗しました。管理者に問い合わせてください。"})
		return
	}
	client := &http.Client{}
	data := map[string]string{ "uid": uid, "code": code }
	body, err := json.Marshal(data)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "response.html", gin.H{"Message": "NFCタグとログインの紐付けに失敗しました。管理者に問い合わせてください。"})
		return
	}
	req, err := http.NewRequest("POST", "http://localhost:4242/receive-uid", strings.NewReader(string(body)))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "response.html", gin.H{"Message": "NFCタグとログインの紐付けに失敗しました。管理者に問い合わせてください。"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + os.Getenv("API_KEY"))
	res, err := client.Do(req)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "response.html", gin.H{"Message": "NFCタグとログインの紐付けに失敗しました。管理者に問い合わせてください。"})
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		c.HTML(http.StatusOK, "response.html", gin.H{"Message": "NFCタグとログインの紐付けが完了しました。ブラウザを閉じてください。"})
	} else if res.StatusCode == 409 {
		c.HTML(http.StatusConflict, "response.html", gin.H{"Message": "既に登録されているログインです。ブラウザを閉じてださい。"})
	} else {
		c.HTML(res.StatusCode, "response.html", gin.H{"Message": "NFCタグとログインの紐付けに失敗しました。管理者に問い合わせてください。"})
	}
}