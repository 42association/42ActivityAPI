package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"42ActivityAPI/internal/accessdb"
)

type M5StickRequestData struct {
	Mac string `json:"mac"`
	RoleName string `json:"role"`
	LocationName string `json:"location"`
}

func AddM5Stick(c *gin.Context) {
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
	if err := AddM5StickToDB(requestData.Mac, requestData.RoleName, requestData.LocationName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mac": requestData.Mac, "role": requestData.RoleName, "location": requestData.LocationName})
	return
}
