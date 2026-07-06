package handler

import (
	"net"
	"net/http"
	"strconv"

	"lmvpn/internal/db"
	"lmvpn/internal/model"
	"lmvpn/internal/vpn"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/gin-gonic/gin"
)

type vpnSettingsResponse struct {
	Enabled             bool   `json:"enabled"`
	Subnet              string `json:"subnet"`
	MTU                 int    `json:"mtu"`
	InterfaceName       string `json:"interface_name"`
	AllowClientToClient bool   `json:"allow_client_to_client"`
	DoLocalIPConfig     bool   `json:"do_local_ip_config"`
	DoRemoteIPConfig    bool   `json:"do_remote_ip_config"`
}

type updateVpnSettingsRequest struct {
	Enabled             *bool   `json:"enabled"`
	Subnet              *string `json:"subnet"`
	MTU                 *int    `json:"mtu"`
	InterfaceName       *string `json:"interface_name"`
	AllowClientToClient *bool   `json:"allow_client_to_client"`
	DoLocalIPConfig     *bool   `json:"do_local_ip_config"`
	DoRemoteIPConfig    *bool   `json:"do_remote_ip_config"`
}

func loadVpnSettings() (model.VpnSetting, error) {
	var s model.VpnSetting
	err := db.DB.First(&s, model.VpnSettingSingletonID).Error
	return s, err
}

func loadReservationsMap() (map[uint]string, error) {
	var rows []model.VpnReservation
	if err := db.DB.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]string, len(rows))
	for _, r := range rows {
		out[r.UserID] = r.IPAddress
	}
	return out, nil
}

func ApplyVpnFromDB(svc *vpn.VpnService) error {
	s, err := loadVpnSettings()
	if err != nil {
		return err
	}
	reservations, err := loadReservationsMap()
	if err != nil {
		return err
	}
	return svc.ApplySettings(s, reservations)
}

func validateSubnet(subnet string) error {
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	if ip.To4() == nil {
		return errIPv4Only
	}
	ones, _ := ipNet.Mask.Size()
	if ones > 30 {
		return errSubnetTooSmall
	}
	return nil
}

var (
	errIPv4Only       = errStr("仅支持 IPv4 子网")
	errSubnetTooSmall = errStr("子网前缀长度不能大于 /30")
	errIPNotInSubnet  = errStr("IP 不在子网范围内")
	errIPReserved     = errStr("该 IP 已被预留")
	errIPIsServer     = errStr("该 IP 为服务器 IP，不可预留")
)

type errStr string

func (e errStr) Error() string { return string(e) }

func GetVpnSettings(c *gin.Context) {
	s, err := loadVpnSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载设置失败"})
		return
	}
	c.JSON(http.StatusOK, vpnSettingsResponse{
		Enabled:             s.Enabled,
		Subnet:              s.Subnet,
		MTU:                 s.MTU,
		InterfaceName:       s.InterfaceName,
		AllowClientToClient: s.AllowClientToClient,
		DoLocalIPConfig:     s.DoLocalIPConfig,
		DoRemoteIPConfig:    s.DoRemoteIPConfig,
	})
}

func UpdateVpnSettings(c *gin.Context) {
	var req updateVpnSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	s, err := loadVpnSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载设置失败"})
		return
	}

	if req.Subnet != nil && *req.Subnet != s.Subnet {
		if err := validateSubnet(*req.Subnet); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		s.Subnet = *req.Subnet
	}
	if req.MTU != nil {
		if *req.MTU < 500 || *req.MTU > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MTU 范围 500-65535"})
			return
		}
		s.MTU = *req.MTU
	}
	if req.InterfaceName != nil {
		s.InterfaceName = *req.InterfaceName
	}
	if req.Enabled != nil {
		s.Enabled = *req.Enabled
	}
	if req.AllowClientToClient != nil {
		s.AllowClientToClient = *req.AllowClientToClient
	}
	if req.DoLocalIPConfig != nil {
		s.DoLocalIPConfig = *req.DoLocalIPConfig
	}
	if req.DoRemoteIPConfig != nil {
		s.DoRemoteIPConfig = *req.DoRemoteIPConfig
	}

	if err := db.DB.Save(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存设置失败"})
		return
	}

	if err := ApplyVpnFromDB(vpn.VPN); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "应用设置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "设置已更新"})
}

type vpnStatusResponse struct {
	Enabled    bool     `json:"enabled"`
	Online     int      `json:"online"`
	UsedIPs    int      `json:"used_ips"`
	Capacity   uint64   `json:"capacity"`
	Clients    []vpn.ClientInfo `json:"clients"`
}

func GetVpnStatus(c *gin.Context) {
	s, err := loadVpnSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载设置失败"})
		return
	}
	used, cap := vpn.VPN.AllocStats()
	clients := vpn.VPN.ClientList()
	c.JSON(http.StatusOK, vpnStatusResponse{
		Enabled:  s.Enabled,
		Online:   len(clients),
		UsedIPs:  used,
		Capacity: cap,
		Clients:  clients,
	})
}

func GetVpnDiag(c *gin.Context) {
	c.JSON(http.StatusOK, vpn.Diag(vpn.VPN))
}

type reservationResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
	CreatedAt string `json:"created_at"`
}

func ListVpnReservations(c *gin.Context) {
	var rows []model.VpnReservation
	if err := db.DB.Order("id asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载预留失败"})
		return
	}
	userIDs := make([]uint, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
	}
	nameMap := make(map[uint]string)
	if len(userIDs) > 0 {
		var users []model.User
		db.DB.Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			nameMap[u.ID] = u.Username
		}
	}
	out := make([]reservationResponse, len(rows))
	for i, r := range rows {
		out[i] = reservationResponse{
			ID:        r.ID,
			UserID:    r.UserID,
			Username:  nameMap[r.UserID],
			IPAddress: r.IPAddress,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	c.JSON(http.StatusOK, gin.H{"reservations": out})
}

type createReservationRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	IPAddress string `json:"ip_address" binding:"required"`
}

func CreateVpnReservation(c *gin.Context) {
	var req createReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在"})
		return
	}

	s, err := loadVpnSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载设置失败"})
		return
	}
	_, ipNet, err := net.ParseCIDR(s.Subnet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "子网配置错误"})
		return
	}
	ip := net.ParseIP(req.IPAddress)
	if ip == nil || !ipNet.Contains(ip) {
		c.JSON(http.StatusBadRequest, gin.H{"error": errIPNotInSubnet.Error()})
		return
	}
	serverIP, _ := cidr.Host(ipNet, 1)
	if ip.Equal(serverIP) {
		c.JSON(http.StatusBadRequest, gin.H{"error": errIPIsServer.Error()})
		return
	}

	var count int64
	db.DB.Model(&model.VpnReservation{}).Where("ip_address = ?", req.IPAddress).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": errIPReserved.Error()})
		return
	}
	var existUser model.VpnReservation
	if err := db.DB.Where("user_id = ?", req.UserID).First(&existUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该用户已有预留 IP"})
		return
	}

	r := model.VpnReservation{UserID: req.UserID, IPAddress: req.IPAddress}
	if err := db.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建预留失败"})
		return
	}

	if vpn.VPN.Running() {
		vpn.VPN.AddReservation(req.UserID, req.IPAddress)
	}
	c.JSON(http.StatusOK, gin.H{"message": "预留已创建"})
}

func DeleteVpnReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var r model.VpnReservation
	if err := db.DB.First(&r, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预留不存在"})
		return
	}
	if err := db.DB.Delete(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	if vpn.VPN.Running() {
		vpn.VPN.RemoveReservation(r.UserID)
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
