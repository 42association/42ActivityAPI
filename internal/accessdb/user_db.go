package accessdb

import (
	"errors"
	"gorm.io/gorm"
)

func AddUsersToDB(users []UserRequestData) ([]string, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, err
	}
	var addedLogin []string

	// 配列を順に処理
	for _, u := range users {
		// loginがDBに存在するか確認
		var existingUser User
		var user User
		if err := db.Where("login = ?", u.Login).First(&existingUser).Error; err != nil {
			// DBに存在しない以外のエラーはエラーとして返す
			if err != gorm.ErrRecordNotFound {
				return addedLogin, err
			}
		} else {
			// DBに存在する場合
			if result := db.Model(&existingUser).Updates(User{UID: u.Uid, Wallet: u.Wallet}); result.Error != nil {
				return addedLogin, result.Error
			}
			addedLogin = append(addedLogin, u.Login)
			continue
		}
		// DBに存在しない場合
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

// Receives uid, login, and wallet, and if the same login does not exist in the DB, adds a new user.(unused)
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