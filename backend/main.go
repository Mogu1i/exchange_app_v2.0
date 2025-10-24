package main

import (
	"exchangeapp/config"
	"exchangeapp/router"
	"io"
	"log"
	"os"
)

func main() {
	//初始化配置信息
	config.InitConfig()
	// fmt.Println(config.Appconfig.App.Port)
	// 创建日志文件
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
	defer logFile.Close()

	// 设置日志输出到文件和控制台
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	//自定义日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("====应用程序启动====")

	//其他配置
	r := router.SetupRouter()
	port := config.Appconfig.App.Port
	if port == "" {
		port = ":8080"
	}
	r.Run(port) // listen and serve on 0.0.0.0:8080
}
