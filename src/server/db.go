package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
	"strconv"
	"github.com/gin-gonic/gin"
	"net/http"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ID, CreatedAt, UpdatedAt are reserved columns in GORM

// User はusersテーブルの行を表す構造体です。
type User struct {
	ID    int
	UID   string
	Login string
}

type Activity struct {
	ID uint `json:"id"`
	UserID int `json:"user_id"`
	M5StickID int `json:"m5stick_id"`
	CreatedAt uint `json:"created_at"`
}

// M5Stick はm5Stickテーブルの行を表す構造体です。
type M5Stick struct {
	ID    int
	Mac   string
	RoleId int
	LocationId int
}

type Location struct {
	ID int
	Name string
}

type Role struct {
	ID int
	Name string
}

func initializeDB() (*gorm.DB, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{}, &M5Stick{}, &Activity{}, &Location{}, &Role{})
	return db, nil	
}

func connectToDB() (*gorm.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// getDSN はDSN（Data Source Name）を環境変数から取得します。
func getDSN() (string, error) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		return "", fmt.Errorf("DB_DSN environment variable is not set")
	}
	return dsn, nil
}

// getCleanData はデータベースから条件に合う掃除データを取得します。/cleanings?start=[UNIXtime]&end=[UNIXtime]
func getCleanData(c *gin.Context, db *gorm.DB) ([]Activity, error) {
	start := c.Query("start")
    end := c.Query("end")

	startInt, err := strconv.ParseInt(start, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start timestamp"})
        return nil, err
    }

    endInt, err := strconv.ParseInt(end, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end timestamp"})
        return nil, err
    }

	if startInt > endInt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range"})
		return nil, err
	}

	rows, err := db.Table("activities").Where("created_at >= ? AND created_at <= ?", startInt, endInt).Rows()
	// rows, err := db.Query("SELECT * FROM activities WHERE timestamp >= ? AND timestamp <= ?", startInt, endInt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "message": err})
        return nil, err
    }
	defer rows.Close()
    // Scan the rows into a slice
    var Activitys []Activity
    for rows.Next() {
		var activity Activity
        err := rows.Scan(&activity.ID, &activity.UserID, &activity.M5StickID, &activity.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row", "message": err})
            return nil, err
		}
		Activitys = append(Activitys, Activity{
			ID: activity.ID,
			UserID: activity.UserID,
			M5StickID: activity.M5StickID,
			CreatedAt: activity.CreatedAt,
		})
    }
	return Activitys, nil
}
