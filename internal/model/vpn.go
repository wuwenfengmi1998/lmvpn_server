package model

import "time"

const VpnSettingSingletonID = 1

type VpnSetting struct {
	ID                  uint   `gorm:"primaryKey"`
	Enabled             bool   `gorm:"default:false"`
	Subnet              string `gorm:"size:64;not null"`
	Subnet6             string `gorm:"size:64"`
	MTU                 int    `gorm:"default:1420"`
	InterfaceName       string `gorm:"size:16"`
	AllowClientToClient bool   `gorm:"default:false"`
	DoLocalIPConfig     bool   `gorm:"default:true"`
	DoRemoteIPConfig    bool   `gorm:"default:true"`
	MaxConnsPerUser     int    `gorm:"default:30"`
	UpdatedAt           time.Time
}

func (VpnSetting) TableName() string {
	return "vpn_settings"
}

type VpnReservation struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	UserID     uint      `gorm:"uniqueIndex;not null"`
	IPAddress  string    `gorm:"size:64;uniqueIndex"`
	IPAddress6 string    `gorm:"size:64"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (VpnReservation) TableName() string {
	return "vpn_reservations"
}

type TrafficStat struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Date      string `gorm:"uniqueIndex;size:10;not null"`
	RxBytes   int64  `gorm:"default:0"`
	TxBytes   int64  `gorm:"default:0"`
	UpdatedAt time.Time
}

func (TrafficStat) TableName() string {
	return "traffic_stats"
}
