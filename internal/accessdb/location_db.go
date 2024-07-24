package accessdb

import (
	"errors"
	"gorm.io/gorm"
)

// Receive the location name, and if it does not exist in the DB, add a new location.
func AddLocationToDB(locationName string) (error, int) {
	db, err := ConnectToDB()
	if err != nil {
		return err, 500
	}

	var existingLocation Location
	if err := db.Where("name = ?", locationName).First(&existingLocation).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err, 500
		}
	} else {
		return errors.New("Location already exists"), 409
	}
	location := Location{Name: locationName}

	if result := db.Create(&location); result.Error != nil {
		return result.Error, 500
	}
	return nil, 200
}
