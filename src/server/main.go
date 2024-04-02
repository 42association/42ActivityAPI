package main

import (
	"encoding/json"
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

func main() {
	router := gin.Default()

	
	// テンプレートエンジンの設定
	router.LoadHTMLGlob("templates/*") // HTMLテンプレートのパスを指定
	
	router.GET("/", func(c *gin.Context) {
		
		// .envファイルの読み込みエラー処理
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
		
		UID := os.Getenv("UID")
			CALLBACK_URL := os.Getenv("CALLBACK_URL")
		// 環境変数が設定されていない場合の処理
		if UID == "" || CALLBACK_URL == "" {
			c.String(http.StatusInternalServerError, "UID or CALLBACK_URL not set")
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"UID":          UID,
			"CALLBACK_URL": CALLBACK_URL,
		})
	})

	router.GET("/callback", func(c *gin.Context) {
		c.HTML(http.StatusOK, "callback.html", nil)
	})

	router.GET("/:uid", redirectToIndexWithUID)

	router.POST("/receive-uid", handleRoot)

	router.Run(":8080")
}

func redirectToIndexWithUID(c *gin.Context) {
	uid := c.Param("uid")
	c.Redirect(http.StatusMovedPermanently, "/?uid="+uid)
}

// リクエストデータの構造を定義
type RequestData struct {
	Code string `json:"code"`
	Uid  string `json:"uid"`
}

type UserData struct {
	IntraName string `json:"intra_name"`
}

// handleRoot関数内でfetchUserDataから返されるintraLoginをレスポンスとして返す
func handleRoot(c *gin.Context) {
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

	// .envファイルの読み込みエラー処理
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	UID := os.Getenv("UID")
	SECRET := os.Getenv("SECRET")
	CALLBACK_URL := os.Getenv("CALLBACK_URL")

	tokenURL := fmt.Sprintf("https://api.intra.42.fr/oauth/token?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		UID, SECRET, code, url.QueryEscape(CALLBACK_URL))

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
