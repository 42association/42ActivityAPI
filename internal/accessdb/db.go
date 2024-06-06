package accessdb

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"errors"
	"net/http"
)

type Shift struct {
	ID	uint   `gorm:"primaryKey"`
	Date  string
	UserID int `json: "user_id"`
	User  User `gorm:"foreignKey:UserID"`
}

type User struct {
	ID    int
	UID   string `gorm:"default:''"`
	Login string
	Wallet string `gorm:"size:42;default:''"`
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

type Date struct {
	Date string
}

type Schedule struct {
	Date string `json:"date"`
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

// Receives the date and returns the shifts for that date.
func GetShiftFromDB(date string) ([]Shift, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, err
	}
	var shifts []Shift
    if err := db.Preload("User").Where("date = ?", date).Find(&shifts).Error; err != nil {
        return nil, err
    }
	return shifts, nil
}

/*
Receives start_time, end_time, and role, and returns activities that were created between start_time,
and end_time and have a matching M5stick role.
*/
func GetActivitiesFromDB(start_time int64, end_time int64, role string) ([]Activity, error) {
	db, err := ConnectToDB()
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

// Receive the role name, and if it does not exist in the DB, add a new role.
func AddRoleToDB(roleName string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var existingRole Role
	if err := db.Where("name = ?", roleName).First(&existingRole).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("Role already exists")
	}
	role := Role{Name: roleName}

	if result := db.Create(&role); result.Error != nil {
		return result.Error
	}
	return nil
}

// Receive the location name, and if it does not exist in the DB, add a new location.
func AddLocationToDB(locationName string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var existingLocation Location
	if err := db.Where("name = ?", locationName).First(&existingLocation).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("Location already exists")
	}
	location := Location{Name: locationName}

	if result := db.Create(&location); result.Error != nil {
		return result.Error
	}
	return nil
}

/*
Receives the MAC address, role name, and location name,
and if the same MAC address does not exist in the DB, adds a new M5stick
*/
func AddM5StickToDB(mac string, roleName string, locationName string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var existingM5Stick M5Stick
	if err := db.Where("mac = ?", mac).First(&existingM5Stick).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("M5Stick already exists")
	}

	var role Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	var location Location
	if err := db.Where("name = ?", locationName).First(&location).Error; err != nil {
		return err
	}

	m5Stick := M5Stick{Mac: mac, RoleId: role.ID, LocationId: location.ID}

	if result := db.Create(&m5Stick); result.Error != nil {
		return result.Error
	}
	return nil
}

// Receives uid, login, and wallet, and if the same login does not exist in the DB, adds a new user.
func AddUserToDB(uid string, login string, wallet string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var existingUser User
	if err := db.Where("login = ?", login).First(&existingUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("User already exists")
	}
	user := User{UID: uid, Login: login, Wallet: wallet}

	if result := db.Create(&user); result.Error != nil {
		return result.Error
	}
	return nil
}

//Receive uid, login, and wallet, and if the same login exists in the DB, update the user data.
func EditUserInDB(uid string, login string, wallet string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var existingUser User
	if err := db.Where("login = ?", login).First(&existingUser).Error; err != nil {
		return err
	}

	if result := db.Model(&existingUser).Updates(User{UID: uid, Wallet: wallet}); result.Error != nil {
		return result.Error
	}
	return nil
}

// Receives the login and returns whether the login exists in the DB.
func UserExists(login string) bool {
	db, err := ConnectToDB()
	if err != nil {
		return false
	}

	var user User
	if err := db.Where("login = ?", login).First(&user).Error; err != nil {
		return false
	}
	return true
}

// Receives the login and uid, and if the login does not have a uid, adds it.
func AddUidToExistUser(login string, uid string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var user User
	if err := db.Where("login = ? AND uid = ?", login, "").First(&user).Error; err != nil {
		return err
	}

	if err := db.Model(&user).Update("uid", uid).Error; err != nil {
		return err
	}
	return nil
}

/*
Receives an array of shifts, adds a shift that does not exist in the DB,
and returns an array of added dates.
*/
func AddShiftToDB(schedule []Schedule) ([]string, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, err
	}

	var addedDate []string
	var flag bool

	for _, s := range schedule {
		if s.Date == "" || len(s.Login)	== 0 {
			continue
		}
		flag = false
		for _, l := range s.Login {
			userId, err := getUserIdFromLogin(db, l)
			if err != nil {
				return nil, err
			}
			var shift Shift
			if err := db.Where("user_id = ? AND date = ?", userId, s.Date).First(&shift).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return nil, err
				}
				shift = Shift{Date: s.Date, UserID: userId}
				if result := db.Create(&shift); result.Error != nil {
					return nil, result.Error
				}
				flag = true
			} else {
				continue
			}
		}
		if flag {
			addedDate = append(addedDate, s.Date)
		}
	}
	return addedDate, nil
}

// Receive the login and *gorm.DB, and return the user ID.
func getUserIdFromLogin(db *gorm.DB, login string) (int, error) {
	var user User
	if err := db.Where("login = ?", login).First(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

// Receive the uid and MAC address, and add a new activity.
func AddActivityToDB(uid string, mac string) (int, string, string, error) {
	db, err := ConnectToDB()
	if err != nil {
		return http.StatusInternalServerError, "", "", err
	}

	var user User
	if err := db.Where("uid = ?", uid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, "", "", err
		} else {
			return http.StatusInternalServerError, "", "", err
		}
	}

	var m5Stick M5Stick
	if err := db.Where("mac = ?", mac).First(&m5Stick).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, "", "", err
		} else {
			return http.StatusInternalServerError, "", "", err
		}
	}

	activity := Activity{UserID: user.ID, M5StickID: m5Stick.ID, CreatedAt: time.Now().Unix()}

	if result := db.Create(&activity); result.Error != nil {
		return http.StatusBadRequest, "", "", result.Error
	}
	return http.StatusOK, uid, mac, nil
}