package logger

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	global *zap.Logger
	once   sync.Once
)

// Init 初始化 zap 日志，日志同时输出到控制台和本地文件。
func Init(dataDir string) (*zap.Logger, error) {
	var initErr error
	once.Do(func() {
		logDir := filepath.Join(dataDir, "logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = err
			return
		}
		logFile, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			initErr = err
			return
		}

		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

		consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
		level := zap.NewAtomicLevelAt(zap.InfoLevel)
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), level),
		)
		global = zap.New(core, zap.AddCaller())
	})
	if initErr != nil {
		return nil, initErr
	}
	return L(), nil
}

// L 返回全局 zap logger。
func L() *zap.Logger {
	if global == nil {
		global = zap.NewNop()
	}
	return global
}
