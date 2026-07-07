package handler

import (
	"net/http"
	"time"

	"lmvpn/internal/db"
	"lmvpn/internal/model"
	"lmvpn/internal/vpn"

	"github.com/gin-gonic/gin"
)

type adminStatsResponse struct {
	UptimeSeconds     int64 `json:"uptime_seconds"`
	ActiveDevices     int   `json:"active_devices"`
	TodayTrafficBytes int64 `json:"today_traffic_bytes"`
	OnlineNodes       int   `json:"online_nodes"`
}

func GetAdminStats(c *gin.Context) {
	var uptime int64
	var onlineNodes int
	if vpn.VPN.Running() {
		uptime = int64(time.Since(vpn.VPN.StartedAt()).Seconds())
		onlineNodes = 1
	}

	activeDevices := len(vpn.VPN.ClientList())

	today := time.Now().Format("2006-01-02")
	var stat model.TrafficStat
	db.DB.Where("date = ?", today).First(&stat)
	liveRx, liveTx := vpn.VPN.TotalLiveTraffic()
	todayTraffic := stat.RxBytes + stat.TxBytes + liveRx + liveTx

	c.JSON(http.StatusOK, adminStatsResponse{
		UptimeSeconds:     uptime,
		ActiveDevices:     activeDevices,
		TodayTrafficBytes: todayTraffic,
		OnlineNodes:       onlineNodes,
	})
}
