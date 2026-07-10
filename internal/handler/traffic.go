package handler

import (
	"net/http"
	"strconv"
	"time"

	"lmvpn/internal/db"
	"lmvpn/internal/model"
	"lmvpn/internal/vpn"

	"github.com/gin-gonic/gin"
)

type userTrafficItem struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	RxBytes    int64  `json:"rx_bytes"`
	TxBytes    int64  `json:"tx_bytes"`
	TotalBytes int64  `json:"total_bytes"`
}

type trafficRecord struct {
	Date    string `json:"date"`
	RxBytes int64  `json:"rx_bytes"`
	TxBytes int64  `json:"tx_bytes"`
}

func parseDays(c *gin.Context) int {
	days := 7
	if d := c.Query("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	return days
}

func GetAdminTrafficToday(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	var stats []model.UserTrafficStat
	db.DB.Where("date = ?", today).Find(&stats)

	userIDs := make([]uint, 0, len(stats))
	for _, s := range stats {
		userIDs = append(userIDs, s.UserID)
	}
	nameMap := make(map[uint]string)
	if len(userIDs) > 0 {
		var users []model.User
		db.DB.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			nameMap[u.ID] = u.Username
		}
	}

	items := make([]userTrafficItem, 0, len(stats))
	var totalRx, totalTx int64
	seen := make(map[uint]bool)
	for _, s := range stats {
		liveRx, liveTx := int64(0), int64(0)
		if vpn.VPN != nil && vpn.VPN.Running() {
			liveRx, liveTx = vpn.VPN.UserLiveTraffic(s.UserID)
		}
		rx := s.RxBytes + liveRx
		tx := s.TxBytes + liveTx
		items = append(items, userTrafficItem{
			UserID:     s.UserID,
			Username:   nameMap[s.UserID],
			RxBytes:    rx,
			TxBytes:    tx,
			TotalBytes: rx + tx,
		})
		totalRx += rx
		totalTx += tx
		seen[s.UserID] = true
	}

	if vpn.VPN != nil && vpn.VPN.Running() {
		for _, ci := range vpn.VPN.ClientList() {
			if seen[ci.UserID] {
				continue
			}
			liveRx, liveTx := vpn.VPN.UserLiveTraffic(ci.UserID)
			if liveRx == 0 && liveTx == 0 {
				continue
			}
			var u model.User
			if err := db.DB.First(&u, ci.UserID).Error; err != nil {
				continue
			}
			items = append(items, userTrafficItem{
				UserID:     ci.UserID,
				Username:   u.Username,
				RxBytes:    liveRx,
				TxBytes:    liveTx,
				TotalBytes: liveRx + liveTx,
			})
			totalRx += liveRx
			totalTx += liveTx
			seen[ci.UserID] = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_rx_bytes": totalRx,
		"total_tx_bytes": totalTx,
		"users":          items,
	})
}

func GetAdminTrafficHistory(c *gin.Context) {
	days := parseDays(c)
	startDate := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	var stats []model.TrafficStat
	db.DB.Where("date >= ?", startDate).Order("date asc").Find(&stats)

	dateMap := make(map[string]trafficRecord, len(stats))
	for _, s := range stats {
		dateMap[s.Date] = trafficRecord{
			Date:    s.Date,
			RxBytes: s.RxBytes,
			TxBytes: s.TxBytes,
		}
	}

	out := make([]trafficRecord, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if r, ok := dateMap[d]; ok {
			out = append(out, r)
		} else {
			out = append(out, trafficRecord{Date: d})
		}
	}

	today := time.Now().Format("2006-01-02")
	var todayStat model.TrafficStat
	db.DB.Where("date = ?", today).First(&todayStat)
	liveRx, liveTx := int64(0), int64(0)
	if vpn.VPN != nil && vpn.VPN.Running() {
		liveRx, liveTx = vpn.VPN.TotalLiveTraffic()
	}

	c.JSON(http.StatusOK, gin.H{
		"today_rx_bytes": todayStat.RxBytes + liveRx,
		"today_tx_bytes": todayStat.TxBytes + liveTx,
		"records":        out,
	})
}

func GetAdminUserTraffic(c *gin.Context) {	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	days := parseDays(c)
	records := queryUserTraffic(uint(id), days)

	today := time.Now().Format("2006-01-02")
	var todayStat model.UserTrafficStat
	db.DB.Where("user_id = ? AND date = ?", id, today).First(&todayStat)
	liveRx, liveTx := int64(0), int64(0)
	if vpn.VPN != nil && vpn.VPN.Running() {
		liveRx, liveTx = vpn.VPN.UserLiveTraffic(uint(id))
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":          user.ID,
		"username":         user.Username,
		"today_rx_bytes":   todayStat.RxBytes + liveRx,
		"today_tx_bytes":   todayStat.TxBytes + liveTx,
		"today_live_rx":    liveRx,
		"today_live_tx":    liveTx,
		"records":          records,
	})
}

func GetMyTrafficToday(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	today := time.Now().Format("2006-01-02")
	var stat model.UserTrafficStat
	db.DB.Where("user_id = ? AND date = ?", uid, today).First(&stat)

	liveRx, liveTx := int64(0), int64(0)
	if vpn.VPN != nil && vpn.VPN.Running() {
		liveRx, liveTx = vpn.VPN.UserLiveTraffic(uid)
	}

	c.JSON(http.StatusOK, gin.H{
		"rx_bytes": stat.RxBytes + liveRx,
		"tx_bytes": stat.TxBytes + liveTx,
	})
}

func GetMyTrafficHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	days := parseDays(c)
	records := queryUserTraffic(uid, days)

	today := time.Now().Format("2006-01-02")
	var todayStat model.UserTrafficStat
	db.DB.Where("user_id = ? AND date = ?", uid, today).First(&todayStat)
	liveRx, liveTx := int64(0), int64(0)
	if vpn.VPN != nil && vpn.VPN.Running() {
		liveRx, liveTx = vpn.VPN.UserLiveTraffic(uid)
	}

	c.JSON(http.StatusOK, gin.H{
		"today_rx_bytes": todayStat.RxBytes + liveRx,
		"today_tx_bytes": todayStat.TxBytes + liveTx,
		"today_live_rx":  liveRx,
		"today_live_tx":  liveTx,
		"records":        records,
	})
}

func queryUserTraffic(userID uint, days int) []trafficRecord {
	startDate := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	var stats []model.UserTrafficStat
	db.DB.Where("user_id = ? AND date >= ?", userID, startDate).Order("date asc").Find(&stats)

	dateMap := make(map[string]trafficRecord, len(stats))
	for _, s := range stats {
		dateMap[s.Date] = trafficRecord{
			Date:    s.Date,
			RxBytes: s.RxBytes,
			TxBytes: s.TxBytes,
		}
	}

	out := make([]trafficRecord, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if r, ok := dateMap[d]; ok {
			out = append(out, r)
		} else {
			out = append(out, trafficRecord{Date: d})
		}
	}
	return out
}
