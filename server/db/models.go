package db

import (
	"github.com/jinzhu/gorm"
)

//Account - user account details
type Account struct {
	gorm.Model
	UserID string `json:"userId"`
	Mobile string `json:"mobile";gorm:"primary_key"`
}
