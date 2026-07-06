package service

import (
	"context"
	"errors"
	"strings"

	"family-score/internal/consts"
	"family-score/internal/logger"
	"family-score/internal/repository"
	"family-score/pkg/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// SetupInit 执行首次初始化，支持家庭组、家长、孩子和默认任务奖励向导。
func (s *Service) SetupInit(ctx context.Context, req SetupInitParam) error {
	logger.L().Info("开始首次初始化", zap.String("admin_login", consts.DefaultAdminLoginName))
	var count int
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("系统已初始化")
	}
	familyName := strings.TrimSpace(req.FamilyName)
	if familyName == "" {
		familyName = "我的家庭"
	}
	adminPassword := strings.TrimSpace(req.AdminPassword)
	if adminPassword == "" {
		adminPassword = consts.DefaultAdminPassword
	}
	if len(adminPassword) < 4 {
		return errors.New("管理员密码至少 4 位")
	}
	groupName := normalizeParentGroup(req.GroupName)
	if groupName == "" {
		groupName = "默认家庭组"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO families(name) VALUES(?)`, familyName)
	if err != nil {
		return err
	}
	familyID, _ := res.LastInsertId()
	if _, err := tx.ExecContext(ctx, `INSERT INTO users(family_id, child_id, display_name, role, login_name, password_hash, enabled) VALUES(?, NULL, '管理员', 'ADMIN', ?, ?, 1)`, familyID, consts.DefaultAdminLoginName, string(hash)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO guardian_groups(family_id, name, description) VALUES(?, ?, ?)`, familyID, groupName, "初始化家庭组"); err != nil {
		return err
	}
	for _, parent := range req.Parents {
		displayName := strings.TrimSpace(parent.DisplayName)
		loginName := strings.TrimSpace(parent.LoginName)
		password := strings.TrimSpace(parent.Password)
		if displayName == "" || loginName == "" || password == "" {
			continue
		}
		if len(password) < 4 {
			return errors.New("家长账号密码至少 4 位")
		}
		parentHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO users(family_id, child_id, display_name, role, parent_title, parent_group, login_name, password_hash, enabled) VALUES(?, NULL, ?, 'PARENT', ?, ?, ?, ?, 1)`, familyID, displayName, normalizeParentTitle(parent.ParentTitle), groupName, loginName, string(parentHash))
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				return errors.New("初始化账号登录名重复")
			}
			return err
		}
	}
	for _, child := range req.Children {
		name := strings.TrimSpace(child.Name)
		if name == "" || child.Age <= 0 {
			continue
		}
		gender := strings.ToUpper(strings.TrimSpace(child.Gender))
		if gender == "" {
			gender = "BOY"
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO children(family_id, parent_group, name, age, gender, profile_note) VALUES(?, ?, ?, ?, ?, ?)`, familyID, groupName, name, child.Age, gender, "初始化孩子档案")
		if err != nil {
			return err
		}
		childID, _ := res.LastInsertId()
		if _, err := tx.ExecContext(ctx, `INSERT INTO score_accounts(child_id, current_month) VALUES(?, ?)`, childID, utils.CurrentMonth()); err != nil {
			return err
		}
		loginName := strings.TrimSpace(child.ChildLoginName)
		password := strings.TrimSpace(child.ChildPassword)
		if loginName != "" || password != "" {
			if loginName == "" || len(password) < 4 {
				return errors.New("孩子账号登录名不能为空，密码至少 4 位")
			}
			childHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO users(family_id, child_id, display_name, role, login_name, password_hash, enabled) VALUES(?, ?, ?, 'CHILD', ?, ?, 1)`, familyID, childID, name, loginName, string(childHash))
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE") {
					return errors.New("初始化孩子账号登录名重复")
				}
				return err
			}
		}
	}
	if req.ImportDefaults || (len(req.Parents) > 0 || len(req.Children) > 0) {
		if err := repository.InsertDefaults(ctx, tx, familyID); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		logger.L().Error("首次初始化事务提交失败", zap.Error(err), zap.Int64("family_id", familyID))
		return err
	}
	logger.L().Info("首次初始化完成", zap.Int64("family_id", familyID), zap.String("admin_login", consts.DefaultAdminLoginName))
	return nil
}
