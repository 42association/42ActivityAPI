package accessdb

import (
	"gorm.io/gorm"
	"errors"
)

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