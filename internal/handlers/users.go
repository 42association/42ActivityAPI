package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"42ActivityAPI/internal/accessdb"
)

type UserRequestData struct {
	Uid string `json:"uid"`
	Login string `json:"login"`
	Wallet string `json:"wallet"`
}

// Handle the endpoint to add a user.
func AddUser(c *gin.Context) {
	var requestData UserRequestData

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Login == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required"})
		return
	}
	if accessdb.UserExists(requestData.Login) {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this login already exists"})
		return
	}
	if err := accessdb.AddUserToDB(requestData.Uid, requestData.Login, requestData.Wallet); err != nil {
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

// Handle the endpoint that updates the user.
func EditUser(c *gin.Context) {
	var requestData UserRequestData

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Login == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required"})
		return
	}
	if err := accessdb.EditUserInDB(requestData.Uid, requestData.Login, requestData.Wallet); err != nil {
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
