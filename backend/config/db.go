// 数据库链接包
package config

import (
	"exchangeapp/global"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	//获取数据库链接需要的信息
	dsn := Appconfig.Database.Dsn
	//进行数据库初始化链接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Faild to initialize database, got error:%v", err)
	}

	//拿到一个通用的数据库对象sqlDB
	sqlDB, err := db.DB()
	//配置打开数据库连接池中空闲链接的最大数量
	sqlDB.SetMaxIdleConns(Appconfig.Database.MaxIdleConns)
	//配置打开数据库连接的最大连接个数
	sqlDB.SetMaxOpenConns(Appconfig.Database.MaxOpenCons)
	//连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		log.Fatalf("Faild to configuredatabase,got error:%v", err)
	}

	global.Db = db
}
