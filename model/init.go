package model

import (
	"log"
	"os"
	"time"

	"go-cloud-disk/conf"
	loglog "go-cloud-disk/utils/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Database 初始化MySQL连接
func Database() {
	connString := conf.MysqlDSN
	// 初始化gorm日志设置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io写入器
		logger.Config{
			SlowThreshold:             time.Second,  // 慢SQL阈值
			LogLevel:                  logger.Error, // 日志级别
			IgnoreRecordNotFoundError: true,         // 忽略ErrRecordNotFound错误日志
			Colorful:                  false,        // 禁用颜色
		},
	)
	// 连接数据库
	db, err := gorm.Open(mysql.Open(connString), &gorm.Config{
		Logger: newLogger,
	})

	if connString == "" || err != nil {
		loglog.Log().Error("MySQL连接失败: %v", err)
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		loglog.Log().Error("MySQL连接失败: %v", err)
		panic(err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	DB = db

	migration()
}
