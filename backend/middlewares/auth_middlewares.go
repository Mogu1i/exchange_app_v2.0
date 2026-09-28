//中间件部分，校验token
//http请求到达我们的处理函数之前，拦截请求并进行相应处理

package middlewares

import (
	"exchangeapp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 优先从 Authorization header 读取（常规请求）
		// 回退到 ?token= query 参数（兼容 EventSource / SSE 请求）
		token := ctx.GetHeader("Authorization")
		if token == "" {
			token = ctx.Query("token")
		}

		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			//处理出现错误时提前结束请求处理,但是会运行完当前函数
			ctx.Abort()
			return
		}

		username, err := utils.ParseJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Set("username", username)
		ctx.Next()
	}

}
