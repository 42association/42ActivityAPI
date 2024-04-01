package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	authorizationEndpoint = "https://api.intra.42.fr/oauth/authorize"
	tokenEndpoint         = "https://api.intra.42.fr/oauth/token"
	clientID              = "u-s4t2ud-xx"
	clientSecret          = "s-s4t2ud-xx"
	redirectURI           = "http://192.168.11.xxx:8080"
	scope                 = "public"
	state                 = "a_very_long_random_string_witchmust_be_unguessable"
)

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func main() {
	http.HandleFunc("/", handleRoot)
	fmt.Println("Server is running on http://192.168.11.xxx:8080")
	http.ListenAndServe(":8080", nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		token := exchangeCodeForToken(code)
		if token != nil {
			fetchUserData(token.AccessToken)
		}
	}
	http.ServeFile(w, r, "src/index.html")
}

func exchangeCodeForToken(code string) *Token {
	tokenURL := fmt.Sprintf("%s?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		tokenEndpoint, clientID, clientSecret, code, url.QueryEscape(redirectURI))

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

func fetchUserData(accessToken string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.intra.42.fr/v2/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching user data: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching user data: %s\n", resp.Status)
		return
	}

	userData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading user data: %v\n", err)
		return
	}

	var userJSON map[string]interface{}
	err = json.Unmarshal(userData, &userJSON)
	if err != nil {
		fmt.Printf("Error parsing user data: %v\n", err)
		return
	}

	intraLogin, ok := userJSON["login"].(string)
	if !ok {
		fmt.Println("Login field not found or not a string")
		return
	}

	fmt.Printf("Intra login: %s\n", intraLogin)
}