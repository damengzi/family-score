package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"family-score/internal/logger"
	"family-score/internal/protocol"
	"family-score/pkg/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type passwordCaptchaSession struct {
	Answer    string
	ExpiresAt time.Time
}

// Login 执行本机账号登录，并限制当天连续输错密码次数。
func (s *Service) Login(ctx context.Context, loginName, password string) (string, protocol.Session, error) {
	loginName = strings.TrimSpace(loginName)
	logger.L().Info("用户尝试登录", zap.String("login_name", loginName))
	var user protocol.User
	var hash string
	var failedDate string
	var failedCount int
	var lockedUntil string
	err := s.repo.DB.QueryRowContext(ctx, `SELECT id, family_id, COALESCE(child_id, 0), role, login_name, display_name, password_hash, COALESCE(failed_login_date, ''), failed_login_count, COALESCE(locked_until, '') FROM users WHERE login_name = ? AND enabled = 1`, loginName).Scan(&user.ID, &user.FamilyID, &user.ChildID, &user.Role, &user.LoginName, &user.Name, &hash, &failedDate, &failedCount, &lockedUntil)
	if err != nil {
		logger.L().Warn("用户登录失败", zap.String("login_name", loginName), zap.Error(err))
		return "", protocol.Session{}, errors.New("账号或密码错误")
	}
	if lockTime, ok := parseLockTime(lockedUntil); ok && time.Now().Before(lockTime) {
		return "", protocol.Session{}, fmt.Errorf("账号已锁定，请在 %s 后再试", lockTime.Format("15:04:05"))
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		msg := s.markLoginFailed(ctx, user.ID, failedDate, failedCount)
		logger.L().Warn("用户密码错误", zap.String("login_name", loginName), zap.Int64("user_id", user.ID))
		return "", protocol.Session{}, errors.New(msg)
	}
	_, _ = s.repo.DB.ExecContext(ctx, `UPDATE users SET failed_login_date = '', failed_login_count = 0, locked_until = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, user.ID)
	token := utils.RandomToken()
	sess := protocol.Session{UserID: user.ID, FamilyID: user.FamilyID, ChildID: user.ChildID, Role: user.Role, Name: user.Name, ExpiresAt: time.Now().Add(12 * time.Hour)}
	s.mu.Lock()
	s.sessions[token] = sess
	s.mu.Unlock()
	logger.L().Info("用户登录成功", zap.Int64("user_id", user.ID), zap.Int64("family_id", user.FamilyID), zap.String("role", user.Role))
	return token, sess, nil
}

func (s *Service) markLoginFailed(ctx context.Context, userID int64, failedDate string, failedCount int) string {
	today := time.Now().Format("2006-01-02")
	count := 1
	if failedDate == today {
		count = failedCount + 1
	}
	if count > 5 {
		lockUntil := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
		_, _ = s.repo.DB.ExecContext(ctx, `UPDATE users SET failed_login_date = ?, failed_login_count = ?, locked_until = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, today, count, lockUntil, userID)
		return "密码错误次数超过 5 次，账号已锁定 15 分钟"
	}
	_, _ = s.repo.DB.ExecContext(ctx, `UPDATE users SET failed_login_date = ?, failed_login_count = ?, locked_until = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, today, count, userID)
	return fmt.Sprintf("账号或密码错误，今日还可尝试 %d 次", 5-count)
}

func parseLockTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// PasswordCaptcha 生成忘记密码图片验证码。
func (s *Service) PasswordCaptcha() protocol.PasswordCaptcha {
	items := []struct{ Key, Name, Emoji string }{{"cat", "小猫", "🐱"}, {"dog", "小狗", "🐶"}, {"rabbit", "兔子", "🐰"}, {"fish", "小鱼", "🐟"}}
	token := utils.RandomToken()
	idx := int(token[0]) % len(items)
	answer := items[idx]
	choices := make([]protocol.PasswordCaptchaChoice, 0, len(items))
	for _, it := range items {
		choices = append(choices, protocol.PasswordCaptchaChoice{Key: it.Key, Image: captchaImage(it.Emoji)})
	}
	s.mu.Lock()
	s.passwordCaptchas[token] = passwordCaptchaSession{Answer: answer.Key, ExpiresAt: time.Now().Add(10 * time.Minute)}
	s.mu.Unlock()
	return protocol.PasswordCaptcha{Token: token, Prompt: "请选择图片中的" + answer.Name, Choices: choices}
}

// ResetPassword 通过图片验证码重置账号密码。
func (s *Service) ResetPassword(ctx context.Context, req protocol.ResetPasswordParam) error {
	req.LoginName = strings.TrimSpace(req.LoginName)
	req.CaptchaAnswer = strings.TrimSpace(req.CaptchaAnswer)
	if req.LoginName == "" || len(req.Password) < 4 {
		return errors.New("登录名不能为空，密码至少 4 位")
	}
	s.mu.Lock()
	captcha, ok := s.passwordCaptchas[req.CaptchaToken]
	if ok {
		delete(s.passwordCaptchas, req.CaptchaToken)
	}
	s.mu.Unlock()
	if !ok || time.Now().After(captcha.ExpiresAt) || captcha.Answer != req.CaptchaAnswer {
		return errors.New("图片验证码错误或已过期")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res, err := s.repo.DB.ExecContext(ctx, `UPDATE users SET password_hash = ?, failed_login_date = '', failed_login_count = 0, locked_until = NULL, updated_at = CURRENT_TIMESTAMP WHERE login_name = ? AND enabled = 1`, string(hash), req.LoginName)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("账号不存在或已注销")
	}
	logger.L().Info("用户通过忘记密码重置密码", zap.String("login_name", req.LoginName))
	return nil
}

func captchaImage(emoji string) string {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="96" height="64" viewBox="0 0 96 64"><rect width="96" height="64" rx="14" fill="#f3f7ff"/><text x="48" y="42" text-anchor="middle" font-size="34">%s</text></svg>`, emoji)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
