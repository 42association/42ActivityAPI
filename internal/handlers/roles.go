package handlers

import (
	"42ActivityAPI/internal/accessdb"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RoleRequestData struct {
	Name string `json:"name"`
}

// Handles the endpoint to add a role.
func AddRole(c *gin.Context) {
	var requestData RoleRequestData

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}
	if err := accessdb.AddRoleToDB(requestData.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": requestData.Name})
	return
}
