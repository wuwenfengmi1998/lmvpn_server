package db

import (
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"lmvpn/internal/config"
	"lmvpn/internal/model"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.DatabaseConfig) error {
	var d gorm.Dialector

	switch cfg.Type {
	case "sqlite":
		d = sqlite.Open(cfg.Path)
	case "mysql":
		if cfg.DSN == "" {
			return fmt.Errorf("mysql DSN 不能为空")
		}
		d = mysql.Open(cfg.DSN)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}

	var err error
	DB, err = gorm.Open(d, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	if err := DB.AutoMigrate(&model.User{}, &model.Session{}, &model.VpnSetting{}, &model.VpnReservation{}); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	if err := seedDefaultVpnSettings(); err != nil {
		return fmt.Errorf("初始化 VPN 设置失败: %w", err)
	}

	if err := seedDefaultAdmin(cfg); err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	log.Printf("数据库初始化完成: %s", cfg.Type)
	return nil
}

func seedDefaultAdmin(cfg *config.DatabaseConfig) error {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	password, err := generateRandomPassword(16)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Username: "admin",
		Password: string(hash),
		Role:     "admin",
		Status:   1,
	}

	if err := DB.Create(admin).Error; err != nil {
		return err
	}

	fmt.Println("========================================")
	fmt.Println("已创建默认管理员账户")
	fmt.Println("用户名: admin")
	fmt.Println("密码: " + password)
	fmt.Println("请登录后立即修改密码！")
	fmt.Println("========================================")

	dbDir := filepath.Dir(cfg.Path)
	pwdFile := filepath.Join(dbDir, ".initial_admin_password")
	if err := os.WriteFile(pwdFile, []byte("admin:"+password+"\n"), 0600); err != nil {
		log.Printf("警告: 写入初始密码文件失败: %v", err)
	} else {
		log.Printf("初始密码已写入 %s，请登录后删除此文件", pwdFile)
	}

	return nil
}

func seedDefaultVpnSettings() error {
	var s model.VpnSetting
	if err := DB.First(&s, model.VpnSettingSingletonID).Error; err == nil {
		return nil
	}
	s = model.VpnSetting{
		ID:               model.VpnSettingSingletonID,
		Enabled:          false,
		Subnet:           "192.168.3.0/24",
		MTU:              1420,
		InterfaceName:    "",
		DoLocalIPConfig:  true,
		DoRemoteIPConfig: true,
	}
	return DB.Create(&s).Error
}

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}
