package main

import (
	"exchangeapp/config"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"log"
)

func main() {
	config.InitConfig()

	if global.Db == nil {
		log.Fatal("❌ 数据库未初始化，请检查 InitConfig() 中的 global.Db 赋值")
	}

	fmt.Println("✅ 数据库初始化成功")

	// 确保表存在
	global.Db.AutoMigrate(&models.Article{})

	const count = 1000
	var articles []models.Article

	for i := 0; i < count; i++ {
		articles = append(articles, models.Article{
			Title:   fmt.Sprintf("性能测试文章 #%d", i+1),
			Content: "这是批量插入测试文章内容",
			Preview: "这是预览部分",
		})
	}

	if err := global.Db.CreateInBatches(articles, 100).Error; err != nil {
		log.Fatalf("❌ 插入失败: %v", err)
	}

	fmt.Printf("✅ 成功插入 %d 条文章\n", count)
}
