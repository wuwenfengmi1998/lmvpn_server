package db

import (
	"fmt"
	"log"

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

	if err := DB.AutoMigrate(&model.User{}, &model.Session{}); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	if err := seedDefaultAdmin(); err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	log.Printf("数据库初始化完成: %s", cfg.Type)
	return nil
}

func seedDefaultAdmin() error {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
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

	log.Println("已创建默认管理员: admin / admin123")
	return nil
}
