package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"family-score/internal/consts"
	"family-score/internal/logger"
	"family-score/internal/protocol"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UpdateUser 修改用户显示名称、孩子账号绑定和登录密码。
func (s *Service) UpdateUser(ctx context.Context, sess protocol.Session, userID int64, req protocol.UpdateUserParam) (protocol.User, error) {
	if !IsAdmin(sess) {
		return protocol.User{}, errors.New("只有管理员可以修改用户")
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" {
		return protocol.User{}, errors.New("显示名称不能为空")
	}
	var current protocol.User
	var passwordHash string
	err := s.repo.DB.QueryRowContext(ctx, `SELECT id, family_id, COALESCE(child_id, 0), role, login_name, display_name, password_hash, enabled, created_at FROM users WHERE id = ? AND family_id = ?`, userID, sess.FamilyID).Scan(&current.ID, &current.FamilyID, &current.ChildID, &current.Role, &current.LoginName, &current.Name, &passwordHash, &current.Enabled, &current.CreatedAt)
	if err != nil {
		return protocol.User{}, errors.New("用户不存在")
	}
	childID := sql.NullInt64{}
	if current.Role == consts.RoleChild {
		if req.ChildID <= 0 {
			return protocol.User{}, errors.New("孩子账号必须绑定孩子档案")
		}
		if !s.CanAccessChild(ctx, sess, req.ChildID) {
			return protocol.User{}, errors.New("无权绑定该孩子档案")
		}
		childID = sql.NullInt64{Int64: req.ChildID, Valid: true}
	}
	newHash := passwordHash
	if strings.TrimSpace(req.Password) != "" {
		if len(req.Password) < 4 {
			return protocol.User{}, errors.New("密码至少 4 位")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return protocol.User{}, err
		}
		newHash = string(hash)
	}
	_, err = s.repo.DB.ExecContext(ctx, `UPDATE users SET display_name = ?, child_id = ?, password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND family_id = ?`, req.DisplayName, childID, newHash, userID, sess.FamilyID)
	if err != nil {
		return protocol.User{}, err
	}
	updated := current
	updated.Name = req.DisplayName
	updated.ChildID = 0
	if childID.Valid {
		updated.ChildID = childID.Int64
	}
	s.refreshUserSessions(updated)
	logger.L().Info("管理员修改用户", zap.Int64("user_id", userID), zap.Int64("operator_id", sess.UserID))
	return updated, nil
}

func (s *Service) refreshUserSessions(user protocol.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, active := range s.sessions {
		if active.UserID == user.ID {
			active.Name = user.Name
			active.ChildID = user.ChildID
			s.sessions[token] = active
		}
	}
}

func (s *Service) clearSessionsByUsersOrChild(userIDs map[int64]struct{}, childID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, active := range s.sessions {
		if _, ok := userIDs[active.UserID]; ok || active.ChildID == childID {
			delete(s.sessions, token)
		}
	}
}
