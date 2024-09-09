package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"short_link/libraries"
	"short_link/models"
	"short_link/routes"
	"strconv"
)

func main() {

	// 日志配置
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	if isDebug, _ := strconv.ParseBool(os.Getenv("DEBUG")); isDebug {
		// 日志颜色化
		gin.ForceConsoleColor()
		// 测试环境日志级别
		logrus.SetLevel(logrus.DebugLevel)
		// 链路追踪（性能损失）
		logrus.SetReportCaller(true)
	}

	r := gin.Default()

	// 加载 env
	envErr := godotenv.Load()
	if envErr != nil {
		panic("Env 配置加载失败：" + envErr.Error())
		return
	}

	// 数据库连接
	models.Conn()
	defer models.DisConn()

	// redis 连接
	libraries.RedisConn(nil)
	defer libraries.RedisDisConn()

	// 路由注册
	routes.Register(r)

	// 8080 端口启动
	err := r.Run(":8080")
	if err != nil {
		panic("http 服务启动失败：" + err.Error())
		return
	}
}
