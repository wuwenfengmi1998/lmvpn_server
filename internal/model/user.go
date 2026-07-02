package model

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Username  string    `gorm:"uniqueIndex;size:64;not null"`
	Password  string    `gorm:"size:128;not null"`
	Role      string    `gorm:"size:16;default:user"`
	Status    int       `gorm:"default:1"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}
