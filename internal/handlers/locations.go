package handlers

import (
	"42ActivityAPI/internal/accessdb"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LocationRequestData struct {
	Name string `json:"name"`
}

// Handles the endpoint that adds a location.
func AddLocation(c *gin.Context) {
	var requestData LocationRequestData

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location is required"})
		return
	}
	if err := accessdb.AddLocationToDB(requestData.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": requestData.Name})
	return
}
