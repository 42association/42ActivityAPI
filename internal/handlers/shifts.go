package handlers

import (
	"42ActivityAPI/internal/accessdb"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"time"
)

type ExchangeData struct {
	login1 string `json:"login1"`
	login2 string `json:"login2"`
	date1  string `json:"date1"`
	date2  string `json:"date2"`
}

// Handle the endpoint that gets the shift.
func GetShiftData(c *gin.Context) {
	date, err := getQueryAboutDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query"})
		return
	}

	shifts, err := accessdb.GetShiftFromDB(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get shift"})
		return
	}
	c.JSON(http.StatusOK, shifts)
}

// Handle the endpoint that adds a shift.
func AddShiftData(c *gin.Context) {
	var schedule []accessdb.Schedule

	if err := c.BindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(schedule) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shift is required"})
		return
	}
	if date, err := accessdb.AddShiftToDB(schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"date": date})
	}
}

// Convert a query with a time format like "2006-01-02" to unix seconds.
func getQueryAboutDate(c *gin.Context) (string, error) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if !isDateStringValid(date) {
		return "", errors.New("Invalid date format. It should be in YYYY/MM/DD format")
	}
	return date, nil
}

func isDateStringValid(date string) bool {
	return regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(date)
}

func ExchangeShiftData(c *gin.Context) {
	var e ExchangeData
	if err := c.BindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.login1 == "" || e.login2 == "" || e.date1 == "" || e.date2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login1, login2, date1, and date2 are required"})
		return
	}
	if !isDateStringValid(e.date1) || !isDateStringValid(e.date2) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. It should be in YYYY/MM/DD format"})
		return
	}
	if shift1, shift2, err := accessdb.ExchangeShiftsOnDB(e.login1, e.login2, e.date1, e.date2); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shifts": []accessdb.Shift{shift1, shift2},
	})
}
