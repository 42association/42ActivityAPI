package accessdb

import (
	"errors"
	"gorm.io/gorm"
)

/*
Receives an array of users and updates the login if it exists in the DB,
or creates a new one if it doesn't. Returns the array of users reflected in the DB.
*/
func AddUsersToDB(users []UserRequestData) ([]string, error) {
	var addedLogin []string
	db, err := ConnectToDB()
	if err != nil {
		return addedLogin, err
	}
	for _, u := range users {
		var user User
		if u.Login == "" {
			continue
		}
		if err := db.Where("login = ?", u.Login).First(&user).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return addedLogin, err
			}
		} else {
			if result := db.Model(&user).Updates(User{UID: u.Uid, Wallet: u.Wallet}); result.Error != nil {
				return addedLogin, result.Error
			}
			addedLogin = append(addedLogin, u.Login)
			continue
		}
		user = User{UID: u.Uid, Login: u.Login, Wallet: u.Wallet}
		if result := db.Create(&user); result.Error != nil {
			return addedLogin, result.Error
		}
		addedLogin = append(addedLogin, u.Login)
	}
	return addedLogin, nil
}

// Receive uid, login, and wallet, and if the same login exists in the DB, update the user data.
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
