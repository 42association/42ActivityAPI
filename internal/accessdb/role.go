package accessdb

import (
	"gorm.io/gorm"
	"errors"
)

// Receive the role name, and if it does not exist in the DB, add a new role.
func AddRoleToDB(roleName string) error {
	db, err := ConnectToDB()
	if err != nil {
		return err
	}

	var existingRole Role
	if err := db.Where("name = ?", roleName).First(&existingRole).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	} else {
		return errors.New("Role already exists")
	}
	role := Role{Name: roleName}

	if result := db.Create(&role); result.Error != nil {
		return result.Error
	}
	return nil
}