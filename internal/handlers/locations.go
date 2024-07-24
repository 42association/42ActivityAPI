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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}
	if requestData.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}
	if err, status := accessdb.AddLocationToDB(requestData.Name); err != nil {
		if status == http.StatusConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflict."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": requestData.Name})
	return
}
