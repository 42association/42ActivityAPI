package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// リクエストデータの構造を定義
type RequestData struct {
	Code string `json:"code"`
	Uid  string `json:"uid"`
}

type RoleRequestData struct {
	Name string `json:"name"`
}

type LocationRequestData struct {
	Name string `json:"name"`
}

type UserData struct {
	IntraName string `json:"intra_name"`
}

type Config struct {
	UID         string
	Secret      string
	CallbackURL string
}

func main() {
	_, err := connectToDB();
	if err != nil {
		log.Println("Failed to initialize database: ", err)
		return
	}

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

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

// 環境変数の読み込み
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

// Uid, Codeを受け取り、intraのユーザー情報を取得してデータベースに保存
// addUserToDBから返されるintraLoginをレスポンスとして返す
func HandleUIDSubmission(c *gin.Context) {
	var requestData RequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestData.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	token := exchangeCodeForToken(requestData.Code)
	if token == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is invalid"})
		return
	}
	userData, err := fetchUserData(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user infomation"})
		return
	}
	if userExists(userData.IntraName) {
		if addUidToExistUser(userData.IntraName, requestData.Uid) == false {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this login is already associated with a uid"})
			return
		}
	} else {
		if err := addUserToDB(requestData.Uid, userData.IntraName, ""); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	response := make(gin.H)

	response["login"] = userData.IntraName
	response["uid"] = requestData.Uid

	c.JSON(http.StatusOK, response)
	return
}

func exchangeCodeForToken(code string) *Token {

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	tokenURL := fmt.Sprintf("https://api.intra.42.fr/oauth/token?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		config.UID, config.Secret, code, url.QueryEscape(config.CallbackURL))

	resp, err := http.PostForm(tokenURL, url.Values{})
	if err != nil {
		fmt.Printf("Error exchanging code for token: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error exchanging code for token: %s\n", resp.Status)
		return nil
	}

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		fmt.Printf("Error decoding token: %v\n", err)
		return nil
	}

	return &token
}

// fetchUserData関数がUserDataを返すように変更
func fetchUserData(accessToken string) (*UserData, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.intra.42.fr/v2/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching user data: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching user data: %s\n", resp.Status)
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading user data: %v\n", err)
		return nil, err
	}

	var userJSON map[string]interface{}
	err = json.Unmarshal(userData, &userJSON)
	if err != nil {
		fmt.Printf("Error parsing user data: %v\n", err)
		return nil, err
	}

	intraName, ok := userJSON["login"].(string)
	if !ok {
		fmt.Println("Login field not found or not a string")
		return nil, fmt.Errorf("login field not found or not a string")
	}

	return &UserData{IntraName: intraName}, nil
}

func addRole(c *gin.Context) {
	var requestData RoleRequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Name == "" {
		// Roleが必須
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}
	// データベースにRoleを追加
	if err := addRoleToDB(requestData.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": requestData.Name})
	return
}

func addLocation(c *gin.Context) {
	var requestData LocationRequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Name == "" {
		// Locationが必須
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location is required"})
		return
	}
	// データベースにLocationを追加
	if err := addLocationToDB(requestData.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": requestData.Name})
	return
}
