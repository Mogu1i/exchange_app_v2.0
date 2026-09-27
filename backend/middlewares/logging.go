package middlewares

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func ensureLogDir() string {
	logDir := filepath.Join(".", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Printf("create log dir failed: %v", err)
		return "."
	}
	return logDir
}

func newFileLogger(fileName string) *log.Logger {
	logDir := ensureLogDir()
	path := filepath.Join(logDir, fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("open log file %s failed: %v", path, err)
		return log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
	return log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}

var accessLogger = newFileLogger("access.log")
var errorLogger = newFileLogger("error.log")

func AccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		accessLogger.Printf(
			"[ACCESS] method=%s path=%s status=%d latency=%s client_ip=%s user_agent=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			c.ClientIP(),
			c.Request.UserAgent(),
		)
	}
}

func RecoveryLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err := fmt.Errorf("panic recovered: %v", recovered)
				errorLogger.Printf("[ERROR] method=%s path=%s error=%v", c.Request.Method, c.Request.URL.Path, err)
				c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}
