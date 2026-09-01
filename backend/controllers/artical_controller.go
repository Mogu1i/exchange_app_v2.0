// package controllers（旧版已注释，保留历史记录）
// ...（省略注释掉的旧代码）

package controllers

import (
	"encoding/json"
	"errors"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var cacheKey = "article"

func CreateArticle(ctx *gin.Context) {
	var article models.Article
	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := global.Db.AutoMigrate(&article); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := global.Db.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 新建文章后清除所有分页缓存
	if global.RedisDB != nil {
		keys, _ := global.RedisDB.Keys("article:page:*").Result()
		if len(keys) > 0 {
			_ = global.RedisDB.Del(keys...).Err()
		}
	}

	ctx.JSON(http.StatusCreated, article)
}

func GetArticles(ctx *gin.Context) {
	/*
		分页 + 旁路缓存模式：
		- 支持 ?page=1&pageSize=20，默认第1页每页20条（最大100）
		- 支持 ?cache=on|off 控制是否使用缓存，默认 on
		- Redis 缓存 key：article:page:{page}:size:{pageSize}
		- 响应包含 total/page/pageSize/totalPages 字段供前端分页组件使用
	*/

	// ── 解析分页参数 ──
	page := 1
	pageSize := 20
	if v, err := strconv.Atoi(ctx.DefaultQuery("page", "1")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(ctx.DefaultQuery("pageSize", "20")); err == nil && v > 0 && v <= 100 {
		pageSize = v
	}
	offset := (page - 1) * pageSize

	useCache := ctx.DefaultQuery("cache", "on")
	pageCacheKey := fmt.Sprintf("article:page:%d:size:%d", page, pageSize)

	// ── 分页结果结构 ──
	type PageResult struct {
		Articles   []models.Article `json:"articles"`
		Total      int64            `json:"total"`
		Page       int              `json:"page"`
		PageSize   int              `json:"pageSize"`
		TotalPages int              `json:"totalPages"`
	}

	// ── 尝试从 Redis 获取当前页缓存 ──
	if useCache == "on" && global.RedisDB != nil {
		if cacheData, err := global.RedisDB.Get(pageCacheKey).Result(); err == nil {
			var cached PageResult
			if err := json.Unmarshal([]byte(cacheData), &cached); err == nil {
				ctx.JSON(http.StatusOK, gin.H{
					"source":     "redis",
					"articles":   cached.Articles,
					"total":      cached.Total,
					"page":       cached.Page,
					"pageSize":   cached.PageSize,
					"totalPages": cached.TotalPages,
				})
				return
			}
		} else if err != redis.Nil {
			// Redis 故障不影响主流程，继续走 DB
		}
	}

	// ── 查询总数 ──
	var total int64
	if err := global.Db.Model(&models.Article{}).Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ── 分页查询 ──
	var articles []models.Article
	if err := global.Db.Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// ── 写入分页缓存（TTL 10 分钟）──
	if useCache == "on" && global.RedisDB != nil {
		result := PageResult{
			Articles: articles, Total: total,
			Page: page, PageSize: pageSize, TotalPages: totalPages,
		}
		if data, err := json.Marshal(result); err == nil {
			_ = global.RedisDB.Set(pageCacheKey, data, time.Minute*10).Err()
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"source":     "mysql",
		"articles":   articles,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

func GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article models.Article
	if err := global.Db.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, article)
}
