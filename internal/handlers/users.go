package handlers

import (
	"42ActivityAPI/internal/accessdb"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Handle the endpoint to add users.
func AddUsers(c *gin.Context) {
	var requestData accessdb.Users

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(requestData.Users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not specified"})
		return
	}
	if addedLogin, err := accessdb.AddUsersToDB(requestData.Users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "users": addedLogin})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"users": addedLogin})
	}
	return
}

// Handle the endpoint that updates the user.
func EditUser(c *gin.Context) {
	var requestData accessdb.UserRequestData

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
