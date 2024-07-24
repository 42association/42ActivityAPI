package accessdb

import (
	"errors"
	"gorm.io/gorm"
)

// Receive the role name, and if it does not exist in the DB, add a new role.
func AddRoleToDB(roleName string) (error, int) {
	db, err := ConnectToDB()
	if err != nil {
		return err, 500
	}

	var existingRole Role
	if err := db.Where("name = ?", roleName).First(&existingRole).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err, 500
		}
	} else {
		return errors.New("Role already exists"), 409
	}
	role := Role{Name: roleName}

	if result := db.Create(&role); result.Error != nil {
		return result.Error, 500
	}
	return nil, 200
}
