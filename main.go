package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()

	// 静的ファイルのルートディレクトリを設定
	router.StaticFile("/", "./src/index.html")

	// UIDをクエリパラメータとしてリダイレクトするエンドポイント
	router.GET("/:uid", redirectToIndexWithUID)

	// 既存のエンドポイントを保持
	router.POST("/adduser/:uid", addUser)

	// サーバーを8080ポートで起動
	router.Run(":8080")
}


// redirectToIndexWithUIDは、クエリパラメータとしてUIDを含むURLにリダイレクトします。
func redirectToIndexWithUID(c *gin.Context) {
	uid := c.Param("uid")
	c.Redirect(http.StatusMovedPermanently, "/?uid="+uid)
}

// addUserは、/adduser/:uid エンドポイントへのPOSTリクエストを処理する関数です。
func addUser(c *gin.Context) {
	uid := c.Param("uid")
	var requestBody struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// コンソールにUIDと名前を出力
	fmt.Printf("{uid: %s, name: %s}\n", uid, requestBody.Name)

	// 正常なレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"status": "User added"})
}
// curl -X POST http://localhost:8080/adduser/123 -H "Content-Type: application/json" -d '{"name":"John Doe"}'
