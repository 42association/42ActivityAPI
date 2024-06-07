package accessdb

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
