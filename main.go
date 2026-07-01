package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"family-score/internal/controller"
	"family-score/internal/logger"
	"family-score/internal/repository"
	"family-score/internal/router"
	"family-score/internal/service"

	"go.uber.org/zap"
)

//go:embed web/*
var staticFiles embed.FS

func main() {
	dataDir := os.Getenv("FAMILY_SCORE_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dataDir = filepath.Join(home, ".family-score")
	}

	addr := os.Getenv("FAMILY_SCORE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	lg, err := logger.Init(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	defer lg.Sync()
	lg.Info("应用启动中", zap.String("addr", addr), zap.String("data_dir", dataDir), zap.Int("pid", os.Getpid()))

	pidPath, err := writePIDFile(dataDir)
	if err != nil {
		lg.Fatal("写入 PID 文件失败", zap.Error(err), zap.String("data_dir", dataDir))
	}
	defer removePIDFile(pidPath, lg)

	repo, err := repository.New(dataDir)
	if err != nil {
		lg.Fatal("数据层初始化失败", zap.Error(err))
	}
	svc := service.New(repo)
	if err := svc.EnsureBuiltinAdmin(context.Background()); err != nil {
		lg.Fatal("固定管理员账号初始化失败", zap.Error(err))
	}
	defer func() {
		if err := svc.Close(); err != nil {
			lg.Error("服务资源关闭失败", zap.Error(err))
			return
		}
		lg.Info("服务资源已关闭")
	}()

	ctrl := controller.New(svc)
	engine := router.New(ctrl, staticFiles)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		lg.Info("家庭德育积分系统已启动", zap.String("url", "http://"+addr), zap.String("pid_file", pidPath))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		lg.Info("收到关闭信号，开始优雅关闭", zap.Int("pid", os.Getpid()))
	case err := <-serverErr:
		if err != nil {
			lg.Fatal("HTTP 服务启动失败", zap.Error(err), zap.String("addr", addr))
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		lg.Error("HTTP 服务优雅关闭失败，执行强制关闭", zap.Error(err))
		if closeErr := httpServer.Close(); closeErr != nil {
			lg.Error("HTTP 服务强制关闭失败", zap.Error(closeErr))
		}
	} else {
		lg.Info("HTTP 服务已优雅关闭")
	}

	if err := <-serverErr; err != nil {
		lg.Error("HTTP 服务退出异常", zap.Error(err))
	}
	lg.Info("应用已退出")
}

func writePIDFile(dataDir string) (string, error) {
	pidPath := filepath.Join(dataDir, "app.pid")
	content := []byte(fmt.Sprintf("%d\n", os.Getpid()))
	return pidPath, os.WriteFile(pidPath, content, 0644)
}

func removePIDFile(pidPath string, lg *zap.Logger) {
	if err := os.Remove(pidPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		lg.Warn("删除 PID 文件失败", zap.Error(err), zap.String("pid_file", pidPath))
		return
	}
	lg.Info("PID 文件已删除", zap.String("pid_file", pidPath))
}
