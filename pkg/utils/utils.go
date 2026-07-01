package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// RandomToken 生成随机十六进制令牌。
func RandomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CurrentMonth 返回当前年月，格式为 yyyy-MM。
func CurrentMonth() string { return time.Now().Format("2006-01") }

// Abs 返回整数绝对值。
func Abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Min 返回两个整数中的较小值。
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SQLiteQuote 对字符串做 SQLite 单引号转义。
func SQLiteQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
