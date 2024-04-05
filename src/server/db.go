package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
	"strconv"
	"github.com/gin-gonic/gin"
	"net/http"
)

// InitializeDatabase はデータベース接続の初期化を行います。
// この関数は外部ファイルから呼び出されることを想定しています。
func InitializeDatabase() (*sql.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}
	db, err := connectToDatabase(dsn)
	if err != nil {
		return nil, err
	}
	if err := verifyConnection(db); err != nil {
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

// connectToDatabase はデータベースへの接続を試みます。
func connectToDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// verifyConnection はデータベースへの接続を確認します。
func verifyConnection(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}
	return nil
}

// InsertUser は新しいユーザーをusersテーブルに挿入します。
func InsertUser(db *sql.DB, uid, login string) error {
	_, err := db.Exec("INSERT INTO users (uid, login) VALUES (?, ?)", uid, login)
	if err != nil {
		return err
	}
	fmt.Println("Inserted a new user into the users table.")
	return nil
}

func GetUserByUid(db *sql.DB, uid string) (User, error) {
	var user User
	query := "SELECT id, uid, login FROM users WHERE uid = ?"
	err := db.QueryRow(query, uid).Scan(&user.ID, &user.UID, &user.Login)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetM5StickByMac(db *sql.DB, mac string) (M5Stick, error) {
	var m5Stick M5Stick
	query := "SELECT id, mac FROM m5sticks WHERE mac = ?"
	err := db.QueryRow(query, mac).Scan(&m5Stick.ID, &m5Stick.Mac)
	if err != nil {
		return m5Stick, err
	}
	return m5Stick, nil
}

func InsertActivity(db *sql.DB, user_id, m5Stick_id int) error {
	ts := time.Now().Unix()
	query := "INSERT INTO activities (user_id, m5stick_id, timestamp) VALUES ($1, $2, $3)"
	_, err := db.Exec(query, user_id, m5Stick_id, ts)
	if err != nil {
		return err
	}
	fmt.Println("Inserted a new activity into the activities table.")
	return nil
}

// GetUsers はusersテーブルから全てのユーザーを取得し、表示します。
func GetUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, uid, login FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	fmt.Println("Users:")
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.UID, &u.Login); err != nil {
			return nil, err
		}
		users = append(users, u)
		fmt.Printf("ID: %d, UID: %s, Login: %s\n", u.ID, u.UID, u.Login)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func getCleanData(c *gin.Context, db *sql.DB) ([]Activity, error) {
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

	rows, err := db.Query("SELECT * FROM activities WHERE timestamp >= ? AND timestamp <= ?", startInt, endInt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "message": err})
        return nil, err
    }
    defer rows.Close()

    // Scan the rows into a slice
    var Activitys []Activity
    for rows.Next() {
		var activity Activity
        err := rows.Scan(&activity.id, &activity.user_id, &activity.m5stick_id, &activity.timestamp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row", "message": err})
            return nil, err
		}
		Activitys = append(Activitys, Activity{
			id: activity.id,
			user_id: activity.user_id,
			m5stick_id: activity.m5stick_id,
			timestamp: activity.timestamp,
		})
    }
	return Activitys, nil
}

// User はusersテーブルの行を表す構造体です。
type User struct {
	ID    int
	UID   string
	Login string
}

// M5Stick はm5Stickテーブルの行を表す構造体です。
type M5Stick struct {
	ID    int
	Mac   string
	RoleId int
	LocationId int
}

// Activity はactivitiesテーブルの行を表す構造体です。
type Activity struct {
	id		 int    `json:"id"`
	user_id   int    `json:"user_id"`
	m5stick_id int `json:"m5stick_id"`
	timestamp int `json:"time_stamp"`
}
