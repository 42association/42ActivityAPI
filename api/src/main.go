package main

import (
	"encoding/json"
	"gorm.io/gorm"
	"errors"
	"fmt"
	"time"
	"regexp"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"strconv"
	"github.com/jinzhu/now"
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

type UserRequestData struct {
	Uid string `json:"uid"`
	Login string `json:"login"`
	Wallet string `json:"wallet"`
}

type ActivityRequestData struct {
	Mac string `json:"mac"`
	Uid string `json:"uid"`
}

type RoleRequestData struct {
	Name string `json:"name"`
}

type LocationRequestData struct {
	Name string `json:"name"`
}

type M5StickRequestData struct {
	Mac string `json:"mac"`
	RoleName string `json:"role"`
	LocationName string `json:"location"`
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
	db, err := initializeDB();
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	seed(db)

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

	router.POST("/activities", addActivity)
	router.GET("/activities/cleanings", getCleanData)

	router.POST("/roles", addRole)

	router.POST("/locations", addLocation)

	router.POST("/m5sticks", addM5Stick)

	router.POST("/users", addUser)
	router.PUT("/users", editUser)

	router.Run(":" + os.Getenv("PORT"))
}

func getShiftData(c *gin.Context) {
	date, err := getQueryAboutDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query"})
		return
	}

	//roleがcleaningのactivityを取得
	shifts, err := getShiftFromDB(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get shift"})
		return
	}
	c.JSON(http.StatusOK, shifts)
}

func getCleanData(c *gin.Context) {
	//start_timeとend_timeを取得
	start_time, end_time, err := getQueryAboutTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query"})
		return
	}

	//roleがcleaningのactivityを取得
	Activities, err := getActivitiesFromDB(start_time, end_time, "cleaning")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activities"})
		return
	}
	c.JSON(http.StatusOK, Activities)
}

func getQueryAboutDate(c *gin.Context) (string, error) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(date) {
		return "", errors.New("Invalid date format. It should be in YYYY/MM/DD format")
	}
	return date, nil
}

func getQueryAboutTime(c *gin.Context) (int64, int64, error) {
	var start_time int64
	var end_time int64
	var err error

	start := c.Query("start")
	if start == "" {
		//startパラメータがない場合は当日の0時0分0秒を取得
		start_time = now.BeginningOfDay().Unix()
	} else {
		start_time, err = strconv.ParseInt(start, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	end := c.Query("end")
	if end == "" {
		//endパラメータがない場合はstart_timeの24時間後を取得
		end_time = start_time + 24*60*60
	} else {
		end_time, err = strconv.ParseInt(end, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	if start_time >= end_time {
		return 0, 0, errors.New("Invalid time range")
	}
	return start_time, end_time, nil
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

			db, err := connectToDB()
			if err != nil {
				log.Fatal("Failed to initialize database:", err)
			}
			// 新しいユーザーを挿入
			db.Create(&User{UID: requestData.Uid, Login: userData.IntraName})
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

func addActivity(c *gin.Context) {
	var requestData ActivityRequestData
	
	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// requestDataが空でないことを確認（MacとUidが非空の文字列）
	if requestData.Mac == "" || requestData.Uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All parameters are required"})
		return
	}
	db, err := connectToDB()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Fatal("Failed to initialize database:", err)
		return
	}
	user := User{}
	if err := db.Where("uid = ?", requestData.Uid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("User not found:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Println("Failed to get user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		}
		return
	}
	m5Stick := M5Stick{}
	if err := db.Where("mac = ?", requestData.Mac).First(&m5Stick).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("M5Stick not found:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "M5Stick not found"})
		} else {
			log.Println("Failed to get M5Stick:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get M5Stick"})
		}
		return
	}
	// Add a new activity
	activity := Activity{UserID: user.ID, M5StickID: m5Stick.ID}
	if result := db.Create(&activity); result.Error != nil {
		log.Fatal("Failed to create activity:", result.Error)
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	// 取得したuserDataを含めてレスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"uid": requestData.Uid, "mac": requestData.Mac})
	return
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

func addM5Stick(c *gin.Context) {
	var requestData M5StickRequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Mac == "" || requestData.RoleName == "" || requestData.LocationName == "" {
		// すべてのパラメータが必須
		c.JSON(http.StatusBadRequest, gin.H{"error": "All parameters are required"})
		return
	}
	// データベースにM5Stickを追加
	if err := addM5StickToDB(requestData.Mac, requestData.RoleName, requestData.LocationName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mac": requestData.Mac, "role": requestData.RoleName, "location": requestData.LocationName})
	return
}

func addUser(c *gin.Context) {
	var requestData UserRequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Login == "" {
		// Loginは必須
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required"})
		return
	}
	if userExists(requestData.Login) {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this login already exists"})
		return
	}
	// データベースにUserを追加
	if err := addUserToDB(requestData.Uid, requestData.Login, requestData.Wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := make(gin.H)

	response["login"] = requestData.Login
	if requestData.Uid != "" {
		response["uid"] = requestData.Uid
	}
	if requestData.Wallet != "" {
		response["wallet"] = requestData.Wallet
	}

	c.JSON(http.StatusOK, response)
	return
}

func editUser(c *gin.Context) {
	var requestData UserRequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Login == "" {
		// Loginは必須
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required"})
		return
	}
	// DB上のUserを編集
	if err := editUserInDB(requestData.Uid, requestData.Login, requestData.Wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := make(gin.H)

	response["login"] = requestData.Login
	if requestData.Uid != "" {
		response["uid"] = requestData.Uid
	}
	if requestData.Wallet != "" {
		response["wallet"] = requestData.Wallet
	}

	c.JSON(http.StatusOK, response)
	return
}
