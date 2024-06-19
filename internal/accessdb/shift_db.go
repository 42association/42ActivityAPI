package accessdb

import (
	"gorm.io/gorm"
	"database/sql"
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

// Receives login and date, exchanges the shift, and returns the exchanged shift.
func ExchangeShiftsOnDB(login1, login2, date1, date2 string) (*Shift, *Shift, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, nil, err
	}
	shift1, shift2, err := transactionExchange(db, login1, login2, date1, date2)
	if err != nil {
		return nil, nil, err
	}
	return shift1, shift2, nil
}

func transactionExchange(db *gorm.DB, login1, login2, date1, date2 string) (*Shift, *Shift, error) {
	var shift1, shift2 Shift
	
	err := db.Transaction(func(tx *gorm.DB) error {
		userId1, err := getUserIdFromLogin(tx, login1)
		if err != nil {
			return err
		}
		userId2, err := getUserIdFromLogin(tx, login2)
		if err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND date = ?", userId1, date1).First(&shift1).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND date = ?", userId2, date2).First(&shift2).Error; err != nil {
			return err
		}
		if err := tx.Model(&shift1).Update("user_id", userId2).Error; err != nil {
			return err
		}
		if err := tx.Model(&shift2).Update("user_id", userId1).Error; err != nil {
			return err
		}
		if err := tx.Preload("User").Where("id = ?", shift1.ID).First(&shift1).Error; err != nil {
			return err
		}
		if err := tx.Preload("User").Where("id = ?", shift2.ID).First(&shift2).Error; err != nil {
			return err
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, nil, err
	}
	return &shift1, &shift2, nil
}

// Receives login and date, deletes the shift, and returns the deleted shift.
func DeleteShiftFromDB(login, date string) (*Shift, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, err
	}
	shift, err := transactionDelete(db, login, date)
	if err != nil {
		return nil, err
	}
	return shift, nil
}

func transactionDelete(db *gorm.DB, login, date string) (*Shift, error) {
	var shift Shift
	err := db.Transaction(func(tx *gorm.DB) error {
		userId, err := getUserIdFromLogin(tx, login)
		if err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND date = ?", userId, date).First(&shift).Error; err != nil {
			return err
		}
		if err := tx.Delete(&shift).Error; err != nil {
			return err
		}
		if err := tx.Preload("User").Unscoped().Where("id = ?", shift.ID).First(&shift).Error; err != nil {
			return err
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	return &shift, nil
}
