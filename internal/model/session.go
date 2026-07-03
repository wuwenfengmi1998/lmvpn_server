package model

import "time"

type Session struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	SessionID string    `gorm:"uniqueIndex;size:36;not null"`
	UserID    uint      `gorm:"index;not null"`
	IP        string    `gorm:"size:64"`
	UserAgent string    `gorm:"size:512"`
	Invalid   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	ExpiresAt time.Time `gorm:"not null"`
}

func (Session) TableName() string {
	return "sessions"
}
