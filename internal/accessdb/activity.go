package accessdb

import (
	"time"
	"net/http"
	"gorm.io/gorm"
)

/*
Receives start_time, end_time, and role, and returns activities that were created between start_time,
and end_time and have a matching M5stick role.
*/
func GetActivitiesFromDB(start_time int64, end_time int64, role string) ([]Activity, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, err
	}
	var activities []Activity
	err = db.
		Preload("User").Preload("M5Stick").Preload("M5Stick.Role").Preload("M5Stick.Location").
		Where("created_at >= ? AND created_at <= ?", start_time, end_time).
		Joins("INNER JOIN m5_sticks ON activities.m5_stick_id = m5_sticks.id INNER JOIN roles ON m5_sticks.role_id = roles.id").
		Where("roles.name = ?", role).
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// Receive the uid and MAC address, and add a new activity.
func AddActivityToDB(uid string, mac string) (int, string, string, error) {
	db, err := ConnectToDB()
	if err != nil {
		return http.StatusInternalServerError, "", "", err
	}

	var user User
	if err := db.Where("uid = ?", uid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, "", "", err
		} else {
			return http.StatusInternalServerError, "", "", err
		}
	}

	var m5Stick M5Stick
	if err := db.Where("mac = ?", mac).First(&m5Stick).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, "", "", err
		} else {
			return http.StatusInternalServerError, "", "", err
		}
	}

	activity := Activity{UserID: user.ID, M5StickID: m5Stick.ID, CreatedAt: time.Now().Unix()}

	if result := db.Create(&activity); result.Error != nil {
		return http.StatusBadRequest, "", "", result.Error
	}
	return http.StatusOK, uid, mac, nil
}