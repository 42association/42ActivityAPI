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
	Login1 string `json:"login1"`
	Login2 string `json:"login2"`
	Date1  string `json:"date1"`
	Date2  string `json:"date2"`
}

type DeleteData struct {
	Login string `json:"login"`
	Date  string `json:"date"`
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
	c.JSON(http.StatusOK, gin.H{"shifts": shifts})
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

// Handle the endpoint that exchanges shifts.
func ExchangeShiftData(c *gin.Context) {
	var e ExchangeData
	if err := c.BindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.Login1 == "" || e.Login2 == "" || e.Date1 == "" || e.Date2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login1, login2, date1, and date2 are required"})
		return
	}
	if !isDateStringValid(e.Date1) || !isDateStringValid(e.Date2) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. It should be in YYYY/MM/DD format"})
		return
	}
	if shift1, shift2, err := accessdb.ExchangeShiftsOnDB(e.Login1, e.Login2, e.Date1, e.Date2); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"shifts": []*accessdb.Shift{shift1, shift2},
		})
	}
}

// Handle the endpoint that deletes a shift.
func DeleteShiftData(c *gin.Context) {
	var d DeleteData
	if err := c.BindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if d.Login == "" || d.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login and date are required"})
		return
	}
	if !isDateStringValid(d.Date) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. It should be in YYYY/MM/DD format"})
		return
	}
	if shift, err := accessdb.DeleteShiftFromDB(d.Login, d.Date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, shift)
	}
}
