package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"encoding/json"
	"net/url"
	"log"
	"io"
	"errors"
	"42ActivityAPI/internal/accessdb"
	"42ActivityAPI/internal/loadconfig"
)

type TokenProperty struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type AuthenticationData struct {
	Code string `json:"code"`
	Uid  string `json:"uid"`
}

/*
Receives the uid and code, and gets user information from intra.
If the user is not registered in the database, registers it and returns the login and uid.
*/
func HandleUIDSubmission(c *gin.Context) {
	var requestData AuthenticationData

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
	intraName, err := fetchUserData(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user infomation"})
		return
	}
	if accessdb.UserExists(intraName) {
		if err := accessdb.AddUidToExistUser(intraName, requestData.Uid); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this login is already associated with a uid"})
			return
		}
	} else {
		if err := accessdb.AddUserToDB(requestData.Uid, intraName, ""); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	response := make(gin.H)

	response["login"] = intraName
	response["uid"] = requestData.Uid

	c.JSON(http.StatusOK, response)
	return
}

// Receive the code and return the access token.
func exchangeCodeForToken(code string) *TokenProperty {
	config, err := loadconfig.LoadConfig()
	if err != nil {
		log.Println("Error: Failed to load configuration: ", err)
		return nil
	}

	query := url.Values{
		"grant_type": []string{"authorization_code"},
		"client_id":  []string{config.UID},
		"client_secret": []string{config.Secret},
		"code": []string{code},
		"redirect_uri": []string{url.QueryEscape(config.CallbackURL)},
	}

	endPointURL := "https://api.intra.42.fr/oauth/token?"

	resp, err := http.PostForm(endPointURL, query)
	if err != nil {
		log.Println("Error: Failed to POST request: ", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: HTTP status is not 200 OK: %s\n", resp.Status)
		return nil
	}

	var tokenProperty TokenProperty
	err = json.NewDecoder(resp.Body).Decode(&tokenProperty)
	if err != nil {
		log.Printf("Error: Failed to decode token properties: %v\n", err)
		return nil
	}

	return &tokenProperty
}

/*
Receive the access token, get the user information
using 42 API, and return the intra name.
*/
func fetchUserData(accessToken string) (string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.intra.42.fr/v2/me", nil)
	req.Header.Set("Authorization", "Bearer " + accessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error: Failed to GET request: ", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: HTTP status is not 200 OK: %s\n", resp.Status)
		return "", errors.New("HTTP status is not 200 OK")
	}

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: Failed to read the response body: %v\n", err)
		return "", err
	}

	var userJSON map[string]interface{}
	err = json.Unmarshal(userData, &userJSON)
	if err != nil {
		log.Printf("Error: Failed to convert to struct: %v\n", err)
		return "", err
	}

	intraName, ok := userJSON["login"].(string)
	if !ok {
		log.Println("Error: Login field does not exist or is not a string")
		return "", errors.New("Login field does not exist or is not a string")
	}

	return intraName, nil
}
