package main

import (
	"github.com/joho/godotenv"
	"short_link/models"
)

// 执行数据库迁移
func main() {
	err := godotenv.Load()
	if err != nil {
		panic("加载数据库配置失败：" + err.Error())
		return
	}
	// 数据库连接
	models.Conn()
	defer models.DisConn()

	dbErr := models.Database().
		Set("gorm:table_options", "ENGINE=InnoDB").
		AutoMigrate(
			&models.Application{},
			&models.UrlStore{},
		)

	if dbErr != nil {
		panic("数据库 Migration 失败：" + dbErr.Error())
		return
	}
}
