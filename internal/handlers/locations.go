package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"42ActivityAPI/internal/accessdb"
)

type LocationRequestData struct {
	Name string `json:"name"`
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