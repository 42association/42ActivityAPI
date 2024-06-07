package accessdb

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

type Shift struct {
	ID     uint `gorm:"primaryKey"`
	Date   string
	UserID int  `json: "user_id"`
	User   User `gorm:"foreignKey:UserID"`
}

type User struct {
	ID     int
	UID    string `gorm:"default:''"`
	Login  string
	Wallet string `gorm:"size:42;default:''"`
}

type Activity struct {
	ID        uint    `json: "id"`
	UserID    int     `json: "user_id"`
	User      User    `gorm:"foreignKey:UserID"`
	M5StickID int     `json: "m5stick_id"`
	M5Stick   M5Stick `gorm:"foreignKey:M5StickID"`
	CreatedAt int64   `json: "created_at"`
}

type M5Stick struct {
	ID         int
	Mac        string
	RoleId     int
	Role       Role `gorm:"foreignKey:RoleId"`
	LocationId int
	Location   Location `gorm:"foreignKey:LocationId"`
}

type Location struct {
	ID   int
	Name string
}

type Role struct {
	ID   int
	Name string
}

type Date struct {
	Date string
}

type Schedule struct {
	Date  string   `json:"date"`
	Login []string `json:"login"`
}

func ConnectToDB() (*gorm.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Shift{}, &User{}, &M5Stick{}, &Activity{}, &Location{}, &Role{})
	return db, nil
}

func getDSN() (string, error) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		return "", fmt.Errorf("DB_DSN environment variable is not set")
	}
	return dsn, nil
}
