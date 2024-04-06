package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
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
	ID uint
	UserID int
	M5StickID int
	CreatedAt time.Time
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

// InitializeDatabase はデータベース接続の初期化を行います。
// この関数は外部ファイルから呼び出されることを想定しています。
func InitializeDatabase() (*gorm.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{})
	db.AutoMigrate(&M5Stick{})
	db.AutoMigrate(&Activity{})
	db.AutoMigrate(&Location{})
	db.AutoMigrate(&Role{})
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
