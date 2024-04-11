package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"errors"
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

	CreatedAt	int64 `json: "created_at"`
}

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

func getDSN() (string, error) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		return "", fmt.Errorf("DB_DSN environment variable is not set")
	}
	return dsn, nil
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

func getActivitiesFromDB(start_time int64, end_time int64, role string) ([]Activity, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	var activities []Activity
	err = db.
		Preload("User").Preload("M5Stick").Preload("M5Stick.Role").Preload("M5Stick.Location").
		Where("created_at >= ? AND created_at <= ?", start_time, end_time).
		Joins("INNER JOIN m5_sticks ON activities.m5_stick_id = m5_sticks.id INNER JOIN roles ON m5_sticks.role_id = roles.id").
		Where("roles.name = ?", role).
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func addRoleToDB(roleName string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	// 同じ名前のRoleがすでに存在するかを確認
	var existingRole Role
	if err := db.Where("name = ?", roleName).First(&existingRole).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("Role already exists")
	}
	role := Role{Name: roleName}

	// データベースにRoleを追加
	if result := db.Create(&role); result.Error != nil {
		return result.Error
	}
	return nil
}

func addLocationToDB(locationName string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	// 同じ名前のLocationがすでに存在するかを確認
	var existingLocation Location
	if err := db.Where("name = ?", locationName).First(&existingLocation).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("Location already exists")
	}
	location := Location{Name: locationName}

	// データベースにLocationを追加
	if result := db.Create(&location); result.Error != nil {
		return result.Error
	}
	return nil
}

func addM5StickToDB(mac string, roleName string, locationName string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	// 同じMACアドレスのM5Stickがすでに存在するかを確認
	var existingM5Stick M5Stick
	if err := db.Where("mac = ?", mac).First(&existingM5Stick).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("M5Stick already exists")
	}
	// roleNameからRoleIdを取得
	var role Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}
	roleId := role.ID

	// locationNameからLocationIdを取得
	var location Location
	if err := db.Where("name = ?", locationName).First(&location).Error; err != nil {
		return err
	}
	locationId := location.ID

	m5Stick := M5Stick{Mac: mac, RoleId: roleId, LocationId: locationId}

	// データベースにM5Stickを追加
	if result := db.Create(&m5Stick); result.Error != nil {
		return result.Error
	}
	return nil
}
