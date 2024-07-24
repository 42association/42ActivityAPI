package accessdb

import (
	"errors"
	"gorm.io/gorm"
)

/*
Receives the MAC address, role name, and location name,
and if the same MAC address does not exist in the DB, adds a new M5stick
*/
func AddM5StickToDB(mac string, roleName string, locationName string) (error, int) {
	db, err := ConnectToDB()
	if err != nil {
		return err, 500
	}

	var existingM5Stick M5Stick
	if err := db.Where("mac = ?", mac).First(&existingM5Stick).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err, 500
		}
	} else {
		return errors.New("M5Stick already exists"), 409
	}

	var role Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("Role does not exist"), 404
		}
		return err, 500
	}

	var location Location
	if err := db.Where("name = ?", locationName).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("Location does not exist"), 404
		}
		return err, 500
	}

	m5Stick := M5Stick{Mac: mac, RoleId: role.ID, LocationId: location.ID}

	if result := db.Create(&m5Stick); result.Error != nil {
		return result.Error, 500
	}
	return nil, 200
}
