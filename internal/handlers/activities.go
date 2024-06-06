package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"log"
	"strconv"
	"github.com/jinzhu/now"
	"errors"
	"42ActivityAPI/internal/accessdb"
)

type ActivityRequestData struct {
	Mac string `json:"mac"`
	Uid string `json:"uid"`
}

func getActivityCleanData(c *gin.Context) {
	//start_timeとend_timeを取得
	start_time, end_time, err := getQueryAboutTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query"})
		return
	}

	//roleがcleaningのactivityを取得
	Activities, err := getActivitiesFromDB(start_time, end_time, "cleaning")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activities"})
		return
	}
	c.JSON(http.StatusOK, Activities)
}

func addActivity(c *gin.Context) {
	var requestData ActivityRequestData
	
	// JSONリクエストボディを解析してrequestDataに格納
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// requestDataが空でないことを確認（MacとUidが非空の文字列）
	if requestData.Mac == "" || requestData.Uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All parameters are required"})
		return
	}
	db, err := connectToDB()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Fatal("Failed to initialize database:", err)
		return
	}
	user := User{}
	if err := db.Where("uid = ?", requestData.Uid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("User not found:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Println("Failed to get user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		}
		return
	}
	m5Stick := M5Stick{}
	if err := db.Where("mac = ?", requestData.Mac).First(&m5Stick).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("M5Stick not found:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "M5Stick not found"})
		} else {
			log.Println("Failed to get M5Stick:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get M5Stick"})
		}
		return
	}
	// Add a new activity
	activity := Activity{UserID: user.ID, M5StickID: m5Stick.ID}
	if result := db.Create(&activity); result.Error != nil {
		log.Fatal("Failed to create activity:", result.Error)
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}
	// 取得したuserDataを含めてレスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"uid": requestData.Uid, "mac": requestData.Mac})
	return
}

func getQueryAboutTime(c *gin.Context) (int64, int64, error) {
	var start_time int64
	var end_time int64
	var err error

	start := c.Query("start")
	if start == "" {
		//startパラメータがない場合は当日の0時0分0秒を取得
		start_time = now.BeginningOfDay().Unix()
	} else {
		start_time, err = strconv.ParseInt(start, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	end := c.Query("end")
	if end == "" {
		//endパラメータがない場合はstart_timeの24時間後を取得
		end_time = start_time + 24*60*60
	} else {
		end_time, err = strconv.ParseInt(end, 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	if start_time >= end_time {
		return 0, 0, errors.New("Invalid time range")
	}
	return start_time, end_time, nil
}