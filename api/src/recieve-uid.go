package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"encoding/json"
	"net/url"
	"log"
	"io"
)

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type RequestData struct {
	Code string `json:"code"`
	Uid  string `json:"uid"`
}

type UserData struct {
	IntraName string `json:"intra_name"`
}

/*
Receives the uid and code, and gets user information from intra.
If the user is not registered in the database, registers it and returns the login and uid.
*/
func HandleUIDSubmission(c *gin.Context) {
	var requestData RequestData

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

/*
Receive the access token, get the user information
using 42 API, and return the intra name.
*/
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
