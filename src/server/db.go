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

type User struct {
	ID    int
	UID   string
	Login string
}

type Activity struct {
	ID			uint `json: "id"`
	
	UserID		int `json: "user_id"`
	User 		User `gorm:"foreignKey:UserID"`

	M5StickID	int `json: "m5stick_id"`
	M5Stick		M5Stick `gorm:"foreignKey:M5StickID"`

	CreatedAt	time.Time `json: "created_at"`
}

type M5Stick struct {
	ID    int
	Mac   string
	
	RoleId int
	Role   Role `gorm:"foreignKey:RoleId"`

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
	db.AutoMigrate(&User{}, &M5Stick{}, &Activity{}, &Location{}, &Role{})
	return db, nil	
}

func seed(db *gorm.DB) error {
	// Create a new user
	users := []User{{UID: "foo", Login: "kakiba"}, {UID: "bar", Login: "tanemura"}}
	for _, user := range users {
		if result := db.Create(&user); result.Error != nil {
			return result.Error
		}
	}

	locations := []Location{{Name: "F1"}, {Name: "F2"}}
	for _, location := range locations {
		if result := db.Create(&location); result.Error != nil {
			return result.Error
		}
	}

	roles := []Role{{Name: "Cleaning"}, {Name: "UsingShower"}}
	for _, role := range roles {
		if result := db.Create(&role); result.Error != nil {
			return result.Error
		}
	}

	m5Sticks := []M5Stick{{Mac: "00:00:00:00:00:00", RoleId: 1, LocationId: 1}, {Mac: "11:11:11:11:11:11", RoleId: 2, LocationId: 2}}
	for _, m5Stick := range m5Sticks {
		if result := db.Create(&m5Stick); result.Error != nil {
			return result.Error
		}
	}

	activities := []Activity{
		{UserID: 1, M5StickID: 1},
		{UserID: 2, M5StickID: 2},
	}
	for _, activity := range activities {
		if result := db.Create(&activity); result.Error != nil {
			return result.Error
		}
	}

	return nil
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
