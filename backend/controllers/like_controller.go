package controllers

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"
	if err := global.RedisDB.Incr(likeKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	// 同步到数据库
	id, _ := strconv.Atoi(articleID)
	if err := global.Db.Model(&models.Article{}).
		Where("id = ?", id).
		Update("likes", gorm.Expr("likes + 1")).Error; err != nil {
		log.Printf("数据库更新失败: %v", err)
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully liked artical"})
}

func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"
	likes, err := global.RedisDB.Get(likeKey).Int()
	if err == redis.Nil {
		var artical models.Article
		if err := global.Db.First(&artical, articleID).Error; err == nil {
			likes = artical.Likes
			//将数据同步到redis中
			if err := global.RedisDB.Set(likeKey, likes, 0).Err(); err != nil {
				log.Printf("redis设置失败:%v", err)
			}
		} else {
			likes = 0
		}
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}

	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}
