package accessdb

import (
	"gorm.io/gorm"
)

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
		if s.Date == "" || len(s.Login) == 0 {
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
