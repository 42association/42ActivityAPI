package handlers

import (
	"42ActivityAPI/internal/accessdb"
	"github.com/gin-gonic/gin"
	"net/http"
)

type M5StickRequestData struct {
	Mac          string `json:"mac"`
	RoleName     string `json:"role"`
	LocationName string `json:"location"`
}

// Handles the endpoint to add the M5stick.
func AddM5Stick(c *gin.Context) {
	var requestData M5StickRequestData

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}
	if requestData.Mac == "" || requestData.RoleName == "" || requestData.LocationName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}
	if err, status := accessdb.AddM5StickToDB(requestData.Mac, requestData.RoleName, requestData.LocationName); err != nil {
		if status == http.StatusConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflict."})
			return
		}
		if status == http.StatusNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mac": requestData.Mac, "role": requestData.RoleName, "location": requestData.LocationName})
	return
}
