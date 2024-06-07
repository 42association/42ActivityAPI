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
	"42ActivityAPI/internal/handlers"
	"42ActivityAPI/internal/accessdb"
	"42ActivityAPI/internal/loadconfig"
	"time"
)

func setupTestDB() *gorm.DB {
	var db *gorm.DB
	var err error
	db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&accessdb.Shift{}, &accessdb.User{}, &accessdb.M5Stick{}, &accessdb.Activity{}, &accessdb.Location{}, &accessdb.Role{})
	return db
}

func TestMain(m *testing.M) {
	db := setupTestDB()
	Seed(db)
	code := m.Run()
	os.Exit(code)
}

func TestGetQueryAboutTime(t *testing.T) {
    req, _ := http.NewRequest("GET", "/activities/cleanings?start=100&end=200", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	start, end, _ := handlers.GetQueryAboutTime(c)
	assert.Equal(t, int64(100), start)
	assert.Equal(t, int64(200), end)
}

func TestLoadConfig(t *testing.T) {
	config, _ := loadconfig.LoadConfig()
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
    router.LoadHTMLGlob("../../web/templates/*")

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

	router.LoadHTMLGlob("../../web/templates/*.html")
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
	db.AutoMigrate(&accessdb.Shift{}, &accessdb.User{}, &accessdb.M5Stick{}, &accessdb.Activity{}, &accessdb.Location{}, &accessdb.Role{})
	return db, nil	
}

func testGetShiftFromDB(date string) ([]accessdb.Shift, error) {
	db, err := testConnectToDB()
	if err != nil {
		return nil, err
	}
	var shifts []accessdb.Shift
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

	var user accessdb.User
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

func Seed(db *gorm.DB) error {
	// Create a new user
	users := []accessdb.User{{UID: "foo", Login: "kakiba", Wallet:"0xA0D9F5854A77D4906906BCEDAAEBB3A39D61165A"}, {UID: "bar", Login: "tanemura", Wallet:"42156DF83404D7833BE3DBDB5D1B367964FDF037"}}
	for _, user := range users {
		if result := db.Create(&user); result.Error != nil {
			return result.Error
		}
	}

	shifts := []accessdb.Shift{{Date: "2024-06-01", UserID: 1}, {Date: "2024-06-02", UserID: 2}}
	for _, shift := range shifts {
		if result := db.Create(&shift); result.Error != nil {
			return result.Error
		}
	}

	locations := []accessdb.Location{{Name: "F1"}, {Name: "F2"}}
	for _, location := range locations {
		if result := db.Create(&location); result.Error != nil {
			return result.Error
		}
	}

	roles := []accessdb.Role{{Name: "Cleaning"}, {Name: "UsingShower"}}
	for _, role := range roles {
		if result := db.Create(&role); result.Error != nil {
			return result.Error
		}
	}

	m5Sticks := []accessdb.M5Stick{{Mac: "00:00:00:00:00:00", RoleId: 1, LocationId: 1}, {Mac: "11:11:11:11:11:11", RoleId: 2, LocationId: 2}}
	for _, m5Stick := range m5Sticks {
		if result := db.Create(&m5Stick); result.Error != nil {
			return result.Error
		}
	}

	activities := []accessdb.Activity{
		{UserID: 1, M5StickID: 1, CreatedAt: time.Now().Unix()},
		{UserID: 2, M5StickID: 2, CreatedAt: time.Now().Unix()},
	}
	for _, activity := range activities {
		if result := db.Create(&activity); result.Error != nil {
			return result.Error
		}
	}

	return nil
}
