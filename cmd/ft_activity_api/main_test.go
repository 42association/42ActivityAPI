package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	_ "modernc.org/sqlite"
	"gorm.io/gorm"
	"net/http/httptest"
	"net/http"
	"os"
	"github.com/42association/42ActivityAPI/internal/handlers"
	"github.com/42association/42ActivityAPI/internal/database"
)

func setupTestDB() *gorm.DB {
	var db *gorm.DB
	var err error
	db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Shift{}, &User{}, &M5Stick{}, &Activity{}, &Location{}, &Role{})
	return db
}

func TestMain(m *testing.M) {
	db := setupTestDB()
	seed(db)
	code := m.Run()
	os.Exit(code)
}

func TestGetQueryAboutTime(t *testing.T) {
    req, _ := http.NewRequest("GET", "/activities/cleanings?start=100&end=200", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	start, end, _ := getQueryAboutTime(c)
	assert.Equal(t, int64(100), start)
	assert.Equal(t, int64(200), end)
}


func TestLoadConfig(t *testing.T) {
	config, _ := LoadConfig()
	assert.Equal(t, os.Getenv("UID"), config.UID)
	assert.Equal(t, os.Getenv("SECRET"), config.Secret)
	assert.Equal(t, os.Getenv("CALLBACK_URL"), config.CallbackURL)
}

type MockConfig struct {
    UID         string
    CallbackURL string
}

// Mock LoadConfig function for testing purposes
func MockLoadConfig() (*MockConfig, error) {
    // Mocked configuration values
    config := &MockConfig{
        UID:         "mockUID",
        CallbackURL: "http://mock-callback-url.com",
    }
    return config, nil
}

func TestShowIndexPage(t *testing.T) {
    router := gin.New()
    router.LoadHTMLGlob("templates/*")

    router.GET("/", ShowIndexPage)

    req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatal(err)
    }

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

}

func TestRedirectToIndexWithUID(t *testing.T) {
	router := gin.New()

	router.GET("/redirect/new", RedirectToIndexWithUID)

	req, err := http.NewRequest("GET", "/redirect/new?uid=123", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)

	expectedRedirectURL := "/?uid=123"
	assert.Equal(t, expectedRedirectURL, w.Header().Get("Location"))
}

func TestShowCallbackPage(t *testing.T) {
	router := gin.New()

	router.LoadHTMLGlob("templates/*.html")
	router.GET("/callback", ShowCallbackPage)

	req, err := http.NewRequest("GET", "/callback", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
}

func testConnectToDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Shift{}, &User{}, &M5Stick{}, &Activity{}, &Location{}, &Role{})
	return db, nil	
}

func testGetShiftFromDB(date string) ([]Shift, error) {
	db, err := testConnectToDB()
	if err != nil {
		return nil, err
	}
	var shifts []Shift
    if err := db.Preload("User").Where("date = ?", date).Find(&shifts).Error; err != nil {
        return nil, err
    }
	return shifts, nil
}

func TestGetShiftUsers(t *testing.T) {
	shifts, _ := testGetShiftFromDB("2024-06-01")
	assert.Equal(t, "kakiba", shifts[0].User.Login)
}

func testUserExists(login string) bool {
	db, err := testConnectToDB()
	if err != nil {
		panic("database error")
	}

	var user User
	if err := db.Where("login = ?", login).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false
		}
		// Handle other errors
		panic("database error")
	}
	return true
}

func TestAddDuplicated(t *testing.T) {
	assert.Equal(t, true, testUserExists("kakiba"))
	assert.Equal(t, false, testUserExists("anonymous"))
}
