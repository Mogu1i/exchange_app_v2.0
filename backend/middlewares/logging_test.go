package middlewares

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEnsureLogDirCreatesDirectory(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd failed: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	logDir := ensureLogDir()
	if _, err := os.Stat(logDir); err != nil {
		t.Fatalf("ensureLogDir should create directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(logDir, "access.log")); err == nil {
		// directory exists and file may not exist yet; this is acceptable
	}
}

func TestAccessLogMiddlewareWritesAccessLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	accessFile, err := os.Create(filepath.Join(tempDir, "access.log"))
	if err != nil {
		t.Fatalf("create access log file failed: %v", err)
	}
	defer accessFile.Close()

	accessLogger = log.New(accessFile, "", 0)
	defer func() {
		accessLogger = log.New(io.Discard, "", 0)
	}()

	r := gin.New()
	r.Use(AccessLogMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d but got %d", http.StatusOK, w.Code)
	}

	if err := accessFile.Sync(); err != nil {
		t.Fatalf("sync access log file failed: %v", err)
	}

	content, err := os.ReadFile(accessFile.Name())
	if err != nil {
		t.Fatalf("read access log file failed: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "method=GET") || !strings.Contains(text, "path=/ping") {
		t.Fatalf("access log missing expected values: %q", text)
	}
}

func TestRecoveryLoggerWritesErrorLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	errorFile, err := os.Create(filepath.Join(tempDir, "error.log"))
	if err != nil {
		t.Fatalf("create error log file failed: %v", err)
	}
	defer errorFile.Close()

	errorLogger = log.New(errorFile, "", 0)
	defer func() {
		errorLogger = log.New(io.Discard, "", 0)
	}()

	r := gin.New()
	r.Use(RecoveryLogger())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d but got %d", http.StatusInternalServerError, w.Code)
	}

	if err := errorFile.Sync(); err != nil {
		t.Fatalf("sync error log file failed: %v", err)
	}

	content, err := os.ReadFile(errorFile.Name())
	if err != nil {
		t.Fatalf("read error log file failed: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "panic recovered") || !strings.Contains(text, "path=/panic") {
		t.Fatalf("error log missing expected values: %q", text)
	}
}
