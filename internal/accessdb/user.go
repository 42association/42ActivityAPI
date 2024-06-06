package accessdb

import (
	"gorm.io/gorm"
	"errors"
)

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