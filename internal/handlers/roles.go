package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/42association/42ActivityAPI/internal/accessdb/accessdb"
)

type RoleRequestData struct {
	Name string `json:"name"`
}

func addRole(c *gin.Context) {
	var requestData RoleRequestData

	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Name == "" {
		// Roleが必須
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}
	// データベースにRoleを追加
	if err := addRoleToDB(requestData.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": requestData.Name})
	return
}