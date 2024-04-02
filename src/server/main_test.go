package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAddUserはaddUser関数のテストを行います。
func TestAddUser(t *testing.T) {
	// Ginのルーターを設定
	router := gin.Default()
	router.POST("/adduser/:uid", addUser)

	// テスト用のHTTPリクエストを作成
	uid := "123"
	jsonStr := []byte(`{"name":"John Doe"}`)
	req, err := http.NewRequest("POST", fmt.Sprintf("/adduser/%s", uid), bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// レスポンスを記録するためのRecorderを作成
	w := httptest.NewRecorder()

	// リクエストをルーターに送信
	router.ServeHTTP(w, req)

	// ステータスコードの検証
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	// レスポンスボディの検証
	expected := `{"status":"User added"}`
	if w.Body.String() != expected {
		t.Errorf("Expected response body to be %s, but got %s", expected, w.Body.String())
	}
}
