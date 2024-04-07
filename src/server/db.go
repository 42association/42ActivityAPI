package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ID, CreatedAt, UpdatedAt are reserved columns in GORM

// ID, CreatedAt, UpdatedAt are reserved columns in GORM

// User はusersテーブルの行を表す構造体です。
type User struct {
	ID    int
	UID   string
	Login string
}

type Activity struct {
	ID			uint
	
	UserID		int
	User 		User `gorm:"foreignKey:UserID"`

	M5StickID	int
	M5Stick		M5Stick `gorm:"foreignKey:M5StickID"`

	CreatedAt	time.Time
}

// M5Stick はm5Stickテーブルの行を表す構造体です。
type M5Stick struct {
	ID    int
	Mac   string
	
	RoleId int
	Role   Role `gorm:"foreignKey:RoleId"`

	LocationId int
	Location   Location `gorm:"foreignKey:LocationId"`
}

type Location struct {
	ID int
	Name string
}

type Role struct {
	ID int
	Name string
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

func connectToDB() (*gorm.DB, error) {
	dsn, err := getDSN()
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

func connectToDB() (*gorm.DB, error) {
	dsn, err := getDSN()
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
