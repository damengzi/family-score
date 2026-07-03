package router

import (
	"embed"
	"net/http"
	"strconv"
	"strings"
	"time"

	"family-score/internal/controller"
	"family-score/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// New 创建 Gin 路由。
func New(ctrl *controller.Controller, staticFiles embed.FS) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(zapGinLogger())
	r.Use(gin.Recovery())

	r.GET("/", serveEmbeddedFile(staticFiles, "web/index.html", "text/html; charset=utf-8"))
	r.GET("/styles.css", serveEmbeddedFile(staticFiles, "web/styles.css", "text/css; charset=utf-8"))
	r.GET("/app.js", serveEmbeddedFile(staticFiles, "web/app.js", "application/javascript; charset=utf-8"))
	r.GET("/core.js", serveEmbeddedFile(staticFiles, "web/core.js", "application/javascript; charset=utf-8"))
	r.GET("/pages/*filepath", serveEmbeddedDirFile(staticFiles, "web/pages", "application/javascript; charset=utf-8"))
	r.Any("/api/*path", func(c *gin.Context) { dispatchAPI(ctrl, c) })
	r.NoRoute(serveEmbeddedFile(staticFiles, "web/index.html", "text/html; charset=utf-8"))
	return r
}

func serveEmbeddedFile(staticFiles embed.FS, name string, contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := staticFiles.ReadFile(name)
		if err != nil {
			logger.L().Error("静态资源读取失败", zap.Error(err), zap.String("file", name))
			c.String(http.StatusNotFound, "文件不存在")
			return
		}
		c.Data(http.StatusOK, contentType, data)
	}
}

func serveEmbeddedDirFile(staticFiles embed.FS, baseDir string, contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath := strings.TrimPrefix(c.Param("filepath"), "/")
		if filePath == "" || strings.Contains(filePath, "..") {
			c.String(http.StatusNotFound, "文件不存在")
			return
		}
		name := baseDir + "/" + filePath
		data, err := staticFiles.ReadFile(name)
		if err != nil {
			logger.L().Error("静态资源读取失败", zap.Error(err), zap.String("file", name))
			c.String(http.StatusNotFound, "文件不存在")
			return
		}
		c.Data(http.StatusOK, contentType, data)
	}
}

func zapGinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", latency),
		}
		if len(c.Errors) > 0 || c.Writer.Status() >= http.StatusInternalServerError {
			logger.L().Error("HTTP 请求失败", append(fields, zap.String("errors", c.Errors.String()))...)
			return
		}
		logger.L().Info("HTTP 请求完成", fields...)
	}
}

func dispatchAPI(ctrl *controller.Controller, c *gin.Context) {
	w, r := c.Writer, c.Request
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	parts := []string{}
	if path != "" {
		parts = strings.Split(path, "/")
	}

	if path == "system/status" && r.Method == http.MethodGet {
		ctrl.SystemStatus(w, r)
		return
	}
	if path == "setup/init" && r.Method == http.MethodPost {
		ctrl.SetupInit(w, r)
		return
	}
	if path == "auth/login" && r.Method == http.MethodPost {
		ctrl.Login(w, r)
		return
	}
	if path == "auth/password-captcha" && r.Method == http.MethodGet {
		ctrl.PasswordCaptcha(w, r)
		return
	}
	if path == "auth/reset-password" && r.Method == http.MethodPost {
		ctrl.ResetPassword(w, r)
		return
	}

	sess, ok := ctrl.RequireSession(w, r)
	if !ok {
		return
	}

	switch {
	case path == "auth/me" && r.Method == http.MethodGet:
		ctrl.AuthMe(w, r, sess)
	case path == "auth/logout" && r.Method == http.MethodPost:
		ctrl.Logout(w, r)
	case path == "profile" && r.Method == http.MethodGet:
		ctrl.Profile(w, r, sess)
	case path == "profile/password" && r.Method == http.MethodPost:
		ctrl.ChangeMyPassword(w, r, sess)
	case path == "users" && r.Method == http.MethodGet:
		ctrl.Users(w, r, sess)
	case path == "users" && r.Method == http.MethodPost:
		ctrl.CreateUser(w, r, sess)
	case len(parts) == 2 && parts[0] == "users" && r.Method == http.MethodPatch:
		ctrl.UpdateUser(w, r, sess, parseID(parts[1]))
	case len(parts) == 2 && parts[0] == "users" && r.Method == http.MethodDelete:
		ctrl.DeleteUser(w, r, sess, parseID(parts[1]))
	case path == "children" && r.Method == http.MethodGet:
		ctrl.Children(w, r, sess)
	case path == "children" && r.Method == http.MethodPost:
		ctrl.CreateChild(w, r, sess)
	case len(parts) == 2 && parts[0] == "children" && r.Method == http.MethodPatch:
		ctrl.UpdateChild(w, r, sess, parseID(parts[1]))
	case len(parts) == 2 && parts[0] == "children" && r.Method == http.MethodDelete:
		ctrl.DeleteChild(w, r, sess, parseID(parts[1]))
	case len(parts) == 3 && parts[0] == "children" && parts[2] == "dashboard" && r.Method == http.MethodGet:
		ctrl.Dashboard(w, r, sess, parseID(parts[1]))
	case len(parts) == 3 && parts[0] == "children" && parts[2] == "score-records" && r.Method == http.MethodGet:
		ctrl.ScoreRecords(w, r, sess, parseID(parts[1]))
	case path == "score-records" && r.Method == http.MethodPost:
		ctrl.CreateScoreRecord(w, r, sess)
	case len(parts) == 4 && parts[0] == "children" && parts[2] == "tasks" && parts[3] == "today" && r.Method == http.MethodGet:
		ctrl.TodayTasks(w, r, sess, parseID(parts[1]))
	case len(parts) == 3 && parts[0] == "tasks" && parts[2] == "submit" && r.Method == http.MethodPost:
		ctrl.SubmitTask(w, r, sess, parseID(parts[1]))
	case len(parts) == 3 && parts[0] == "tasks" && parts[2] == "audit" && r.Method == http.MethodPost:
		ctrl.AuditTask(w, r, sess, parseID(parts[1]))
	case path == "task-templates" && r.Method == http.MethodGet:
		ctrl.TaskTemplates(w, r, sess)
	case path == "task-templates" && r.Method == http.MethodPost:
		ctrl.CreateTaskTemplate(w, r, sess)
	case len(parts) == 2 && parts[0] == "task-templates" && r.Method == http.MethodDelete:
		ctrl.DeleteTaskTemplate(w, r, sess, parseID(parts[1]))
	case path == "rewards" && r.Method == http.MethodGet:
		ctrl.Rewards(w, r, sess)
	case path == "rewards" && r.Method == http.MethodPost:
		ctrl.CreateReward(w, r, sess)
	case len(parts) == 2 && parts[0] == "rewards" && r.Method == http.MethodDelete:
		ctrl.DeleteReward(w, r, sess, parseID(parts[1]))
	case path == "exchange-orders" && r.Method == http.MethodGet:
		ctrl.ExchangeOrders(w, r, sess)
	case path == "exchange-orders" && r.Method == http.MethodPost:
		ctrl.CreateExchangeOrder(w, r, sess)
	case len(parts) == 3 && parts[0] == "exchange-orders" && parts[2] == "audit" && r.Method == http.MethodPost:
		ctrl.AuditExchangeOrder(w, r, sess, parseID(parts[1]))
	case path == "system/backup" && r.Method == http.MethodPost:
		ctrl.Backup(w, r, sess)
	case path == "system/backups" && r.Method == http.MethodGet:
		ctrl.Backups(w, r, sess)
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
	}
}

func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}
