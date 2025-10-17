package model

import "github.com/cccvno1/nova/pkg/database"

type User struct {
	database.Model
	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password string `gorm:"not null;size:255" json:"-"`
	Nickname string `gorm:"size:50" json:"nickname"`
	Avatar   string `gorm:"size:255" json:"avatar"`
	Status   int    `gorm:"default:1;not null" json:"status"` // 1: active, 2: disabled
}

func (User) TableName() string {
	return "users"
}
