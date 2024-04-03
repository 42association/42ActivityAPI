package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

type ActivityRequestData struct {
	Mac string `json:"mac"`
	Uid string `json:"uid"`
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
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", ShowIndexPage)
	router.GET("/:uid", RedirectToIndexWithUID)
	router.GET("/callback", ShowCallbackPage)
	router.POST("/receive-uid", HandleUIDSubmission)
	router.POST("/activity/add", addActivity)

	router.Run(":8000")
}

// 環境変数の読み込み
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
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
	uid := c.Param("uid")
	c.Redirect(http.StatusMovedPermanently, "/?uid="+uid)
}

func ShowCallbackPage(c *gin.Context) {
	c.HTML(http.StatusOK, "callback.html", nil)
}

// handleRoot関数内でfetchUserDataから返されるintraLoginをレスポンスとして返す
func HandleUIDSubmission(c *gin.Context) {
	var requestData RequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// requestDataが空でないことを確認（CodeとUidが非空の文字列）
	if requestData.Code != "" && requestData.Uid != "" {
		fmt.Printf("Code: %s, Uid: %s\n", requestData.Code, requestData.Uid)
		token := exchangeCodeForToken(requestData.Code)
		if token != nil {
			userData, err := fetchUserData(token.AccessToken)
			if err != nil {
				// エラーハンドリング
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
				return
			}

			db, err := InitializeDatabase()
			if err != nil {
				log.Fatal("Failed to initialize database:", err)
			}
			defer db.Close()
		
			// 新しいユーザーを挿入
			if err := InsertUser(db, requestData.Uid, userData.IntraName); err != nil {
				log.Fatalf("Failed to insert user: %v", err)
			}
			// 取得したuserDataを含めてレスポンスを返す
			c.JSON(http.StatusOK, gin.H{
				// "code": requestData.Code,
				"uid": requestData.Uid, "intra_login": userData.IntraName})
			return
		}
		c.JSON(http.StatusOK, nil)
	} else {
		// パラメータが空の場合はnullを返す
		c.JSON(http.StatusOK, nil)
	}
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

	userData, err := ioutil.ReadAll(resp.Body)
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

func addActivity(c *gin.Context) {
	var requestData ActivityRequestData
	var user User
	var m5stick M5Stick
	
	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// requestDataが空でないことを確認（MacとUidが非空の文字列）
	if requestData.Mac == "" || requestData.Uid == "" {
		// パラメータが空の場合はnullを返す
		c.JSON(http.StatusOK, nil)
		return
	}
	db, err := InitializeDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	user := GetUserByUid(db, requestData.Uid)
	m5Stick := GetM5StickByMac(db, requestData.Mac)
	// Add a new activity
	if err := InsertActivity(db, m5Stick.id, user.id); err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}
	// 取得したuserDataを含めてレスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"uid": requestData.Uid, "mac": requestData.Mac})
	return
}
