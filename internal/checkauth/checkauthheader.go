package checkauth

import (
	"os"
	"net/http"
	"github.com/gin-gonic/gin"
)
// Validate the API key by matching the environment variables and the Authorization header.
func CheckAuthHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedAPIKey := "Bearer " + os.Getenv("API_KEY")
		if expectedAPIKey == "Bearer " {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error."})
			c.Abort()
			return
		}
		apiKey := c.GetHeader("Authorization")
		if apiKey != expectedAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}