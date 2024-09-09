package models

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"time"
)

// 全局数据库连接对象
var db *gorm.DB

// Conn 初始化数据库连接
func Conn() {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
		"utf8mb4",
		"Local",
	)

	// 建立连接
	var err error
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败：" + err.Error())
	}

	// 连接池设置
	sqlDB, poolErr := db.DB()
	if poolErr != nil {
		panic("数据库连接池初始化失败：" + poolErr.Error())
	}

	// 空闲连接池中的最大连接数
	sqlDB.SetMaxIdleConns(10)
	// 数据库的最大打开连接数。
	sqlDB.SetMaxOpenConns(100)
	// 连接重复使用时长
	sqlDB.SetConnMaxLifetime(time.Hour)
}

// DisConn 连接池手动断开
func DisConn() {
	sqlDB, poolErr := db.DB()
	if poolErr != nil {
		panic("数据库连接池初始化失败：" + poolErr.Error())
	}

	err := sqlDB.Close()
	if err != nil {
		panic("数据库连接池关闭失败：" + poolErr.Error())
	}
}

// Database 获取 mysql 连接
func Database() *gorm.DB {
	if db == nil {
		Conn()
	}
	return db
}
