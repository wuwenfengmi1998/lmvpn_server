package handler

import (
	"net/http"
	"strconv"
	"time"

	"lmvpn/internal/db"
	"lmvpn/internal/model"
	"lmvpn/internal/vpn"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
	Status   *int   `json:"status"`
}

type updateUserRequest struct {
	Status   *int   `json:"status"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func formatUser(u *model.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

var validRoles = map[string]bool{"admin": true, "user": true}

func isValidRole(role string) bool {
	return validRoles[role]
}

func GetUserCount(c *gin.Context) {
	var count int64
	db.DB.Model(&model.User{}).Count(&count)
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func ListUsers(c *gin.Context) {
	var users []model.User
	db.DB.Order("id asc").Find(&users)

	result := make([]userResponse, len(users))
	for i, u := range users {
		result[i] = formatUser(&u)
	}

	c.JSON(http.StatusOK, gin.H{"users": result})
}

func CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var exist model.User
	if err := db.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := model.User{
		Username: req.Username,
		Password: string(hash),
		Role:     "user",
		Status:   1,
	}
	if req.Role != "" {
		if !isValidRole(req.Role) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "角色无效，仅支持 admin 或 user"})
			return
		}
		user.Role = req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	c.JSON(http.StatusOK, formatUser(&user))
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	currentUserID, _ := c.Get("user_id")

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	updates := map[string]interface{}{}

	if req.Status != nil {
		if user.ID == currentUserID.(uint) && *req.Status != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能禁用自己的账号"})
			return
		}
		updates["status"] = *req.Status
	}

	if req.Role != "" {
		if !isValidRole(req.Role) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "角色无效，仅支持 admin 或 user"})
			return
		}
		if user.ID == currentUserID.(uint) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能修改自己的角色"})
			return
		}
		if req.Role != "admin" && user.Role == "admin" {
			var adminCount int64
			db.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount)
			if adminCount <= 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "系统至少需要保留一个管理员"})
				return
			}
		}
		updates["role"] = req.Role
	}

	if req.Password != "" {
		if err := validatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		updates["password"] = string(hash)
		updates["token_invalid_before"] = time.Now()
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有需要更新的字段"})
		return
	}

	if err := db.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	if req.Password != "" || req.Role != "" || req.Status != nil {
		db.DB.Model(&model.Session{}).Where("user_id = ?", id).Update("invalid", true)
		if vpn.VPN != nil && vpn.VPN.Running() {
			vpn.VPN.KickUser(uint(id))
		}
	}

	db.DB.First(&user, id)
	c.JSON(http.StatusOK, formatUser(&user))
}

func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	currentUserID, _ := c.Get("user_id")

	if uint(id) == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}

	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if user.Role == "admin" {
		var adminCount int64
		db.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount)
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "系统至少需要保留一个管理员"})
			return
		}
	}

	if err := db.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
