package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"family-score/internal/consts"
	"family-score/internal/logger"
	"family-score/internal/protocol"
	"family-score/internal/repository"
	"family-score/pkg/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service 表示服务层，负责业务规则编排。
type Service struct {
	repo             *repository.Repository
	sessions         map[string]protocol.Session
	passwordCaptchas map[string]passwordCaptchaSession
	listenAddr       string
	mu               sync.Mutex
}

// New 创建服务层实例。
func New(repo *repository.Repository) *Service {
	logger.L().Info("服务层初始化完成")
	return &Service{repo: repo, sessions: map[string]protocol.Session{}, passwordCaptchas: map[string]passwordCaptchaSession{}}
}

// Close 关闭服务依赖资源。
func (s *Service) Close() error { return s.repo.Close() }

// DataDir 返回本机数据目录。
func (s *Service) DataDir() string { return s.repo.DataDir }

// DBPath 返回 SQLite 文件路径。
func (s *Service) DBPath() string { return s.repo.DBPath }

// SetListenAddr 记录服务实际监听地址，供状态和网络信息查询使用。
func (s *Service) SetListenAddr(addr string) { s.listenAddr = addr }

// ListenAddr 返回服务实际监听地址。
func (s *Service) ListenAddr() string {
	if s.listenAddr == "" {
		return "127.0.0.1:8080"
	}
	return s.listenAddr
}

type SystemStatusParam = protocol.SystemStatusParam
type SetupInitParam = protocol.SetupInitParam
type CreateUserParam = protocol.CreateUserParam
type CreateChildParam = protocol.CreateChildParam
type SelfRegisterParam = protocol.SelfRegisterParam
type CreateScoreRecordParam = protocol.CreateScoreRecordParam
type SubmitTaskParam = protocol.SubmitTaskParam
type AuditParam = protocol.AuditParam
type CreateTaskTemplateParam = protocol.CreateTaskTemplateParam
type CreateRewardParam = protocol.CreateRewardParam
type CreateExchangeOrderParam = protocol.CreateExchangeOrderParam

// SystemStatus 查询本机系统状态。
func (s *Service) SystemStatus(ctx context.Context) (SystemStatusParam, error) {
	var userCount int
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
		return SystemStatusParam{}, err
	}
	return SystemStatusParam{SetupCompleted: userCount > 0, DataDir: s.repo.DataDir, DBPath: s.repo.DBPath, Addr: s.ListenAddr(), Now: time.Now().Format(time.RFC3339)}, nil
}

// SelfRegister 自主注册普通用户。
func (s *Service) SelfRegister(ctx context.Context, req SelfRegisterParam) (int64, error) {
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.LoginName = strings.TrimSpace(req.LoginName)
	if req.DisplayName == "" || req.LoginName == "" || len(req.Password) < 4 {
		return 0, errors.New("用户名、登录名不能为空，密码至少 4 位")
	}
	var familyID int64
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT id FROM families ORDER BY id LIMIT 1`).Scan(&familyID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("请先完成系统初始化，再进行注册")
		}
		return 0, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO users(family_id, display_name, role, login_name, password_hash, enabled) VALUES(?, ?, 'NORMAL', ?, ?, 1)`, familyID, req.DisplayName, req.LoginName, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, errors.New("登录名已存在")
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("自主注册普通用户", zap.Int64("user_id", id), zap.String("login_name", req.LoginName), zap.Int64("family_id", familyID))
	return id, nil
}

// EnsureBuiltinAdmin 确保已有数据也拥有管理员账号 admin。
func (s *Service) EnsureBuiltinAdmin(ctx context.Context) error {
	var familyID int64
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT id FROM families ORDER BY id LIMIT 1`).Scan(&familyID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	var count int
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE login_name = 'admin'`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		_, err := s.repo.DB.ExecContext(ctx, `UPDATE users SET role = 'ADMIN', enabled = 1, child_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE login_name = 'admin'`)
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(consts.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.repo.DB.ExecContext(ctx, `INSERT INTO users(family_id, child_id, display_name, role, login_name, password_hash, enabled) VALUES(?, NULL, '管理员', 'ADMIN', 'admin', ?, 1)`, familyID, string(hash))
	if err == nil {
		logger.L().Info("已补充固定管理员账号", zap.String("login_name", consts.DefaultAdminLoginName), zap.Int64("family_id", familyID))
	}
	return err
}

// Logout 清除本机会话。
func (s *Service) Logout(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// IsAdmin 判断会话是否为管理员。
func IsAdmin(sess protocol.Session) bool {
	return sess.Role == consts.RoleAdmin
}

// CanOperate 判断会话是否拥有业务操作权限。
func CanOperate(sess protocol.Session) bool {
	return sess.Role == consts.RoleAdmin || sess.Role == consts.RoleParent
}

// Users 查询家庭用户列表，管理员可查看全部，家长可查看可归属的家长账号。
func (s *Service) Users(ctx context.Context, sess protocol.Session) ([]protocol.User, error) {
	if !CanOperate(sess) {
		return nil, errors.New("只有管理员或家长可以查看用户")
	}
	query := `SELECT id, family_id, COALESCE(child_id, 0), role, login_name, display_name, COALESCE(parent_title, ''), COALESCE(parent_group, ''), enabled, created_at FROM users WHERE family_id = ? ORDER BY id DESC`
	args := []any{sess.FamilyID}
	if sess.Role == consts.RoleParent {
		parentGroup := s.parentGroupOf(ctx, sess.FamilyID, sess.UserID)
		query = `SELECT id, family_id, COALESCE(child_id, 0), role, login_name, display_name, COALESCE(parent_title, ''), COALESCE(parent_group, ''), enabled, created_at FROM users WHERE family_id = ? AND role = 'PARENT' AND enabled = 1 AND (id = ? OR (? <> '' AND parent_group = ?)) ORDER BY id DESC`
		args = append(args, sess.UserID, parentGroup, parentGroup)
	}
	rows, err := s.repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.User{}
	for rows.Next() {
		var it protocol.User
		if err := rows.Scan(&it.ID, &it.FamilyID, &it.ChildID, &it.Role, &it.LoginName, &it.Name, &it.ParentTitle, &it.ParentGroup, &it.Enabled, &it.CreatedAt); err == nil {
			items = append(items, it)
		}
	}
	return items, nil
}

// CreateUser 注册新用户，仅管理员可用。
func (s *Service) CreateUser(ctx context.Context, sess protocol.Session, req CreateUserParam) (int64, error) {
	if !IsAdmin(sess) {
		return 0, errors.New("只有管理员可以注册用户")
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.LoginName = strings.TrimSpace(req.LoginName)
	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	if req.Role == "" {
		req.Role = consts.RoleChild
	}
	if req.Role != consts.RoleParent && req.Role != consts.RoleChild {
		return 0, errors.New("角色只能是 PARENT 或 CHILD")
	}
	if req.Role == consts.RoleChild && req.ChildID <= 0 {
		return 0, errors.New("孩子账号必须绑定孩子档案")
	}
	if req.Role == consts.RoleChild && !s.CanAccessChild(ctx, sess, req.ChildID) {
		return 0, errors.New("无权绑定该孩子档案")
	}
	parentTitle := ""
	parentGroup := ""
	if req.Role == consts.RoleParent {
		parentTitle = normalizeParentTitle(req.ParentTitle)
		parentGroup = normalizeParentGroup(req.ParentGroup)
		if req.DisplayName == "" {
			req.DisplayName = parentTitle
		}
	}
	if req.DisplayName == "" || req.LoginName == "" || len(req.Password) < 4 {
		return 0, errors.New("用户名、登录名不能为空，密码至少 4 位")
	}
	if err := s.ensureGuardianGroup(ctx, sess.FamilyID, parentGroup); err != nil {
		return 0, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	childID := sql.NullInt64{Int64: req.ChildID, Valid: req.Role == consts.RoleChild}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO users(family_id, child_id, display_name, role, parent_title, parent_group, login_name, password_hash, enabled) VALUES(?, ?, ?, ?, ?, ?, ?, ?, 1)`, sess.FamilyID, childID, req.DisplayName, req.Role, parentTitle, parentGroup, req.LoginName, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, errors.New("登录名已存在")
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("管理员新增用户", zap.Int64("user_id", id), zap.String("login_name", req.LoginName), zap.String("role", req.Role), zap.Int64("operator_id", sess.UserID))
	return id, nil
}

// DeleteUser 注销用户，仅管理员可用。
func (s *Service) DeleteUser(ctx context.Context, sess protocol.Session, userID int64) error {
	if !IsAdmin(sess) {
		return errors.New("只有管理员可以注销用户")
	}
	if userID == sess.UserID {
		return errors.New("不能注销当前登录用户")
	}
	res, err := s.repo.DB.ExecContext(ctx, `UPDATE users SET enabled = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND family_id = ?`, userID, sess.FamilyID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("用户不存在")
	}
	s.mu.Lock()
	for token, active := range s.sessions {
		if active.UserID == userID {
			delete(s.sessions, token)
		}
	}
	s.mu.Unlock()
	logger.L().Info("管理员注销用户", zap.Int64("user_id", userID), zap.Int64("operator_id", sess.UserID))
	return nil
}

// Session 根据 token 获取会话。
func (s *Service) Session(token string) (protocol.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[token]
	if ok && time.Now().After(sess.ExpiresAt) {
		delete(s.sessions, token)
		ok = false
	}
	return sess, ok
}

// Children 查询家庭孩子列表。
func (s *Service) Children(ctx context.Context, sess protocol.Session) ([]protocol.Child, error) {
	query := `SELECT id, family_id, COALESCE(parent_user_id, 0), COALESCE(parent_group, ''), name, age, gender FROM children WHERE family_id = ? ORDER BY id`
	args := []any{sess.FamilyID}
	switch sess.Role {
	case consts.RoleParent:
		parentGroup := s.parentGroupOf(ctx, sess.FamilyID, sess.UserID)
		query = `SELECT id, family_id, COALESCE(parent_user_id, 0), COALESCE(parent_group, ''), name, age, gender FROM children WHERE family_id = ? AND (parent_user_id = ? OR (? <> '' AND parent_group = ?)) ORDER BY id`
		args = append(args, sess.UserID, parentGroup, parentGroup)
	case consts.RoleChild:
		query = `SELECT id, family_id, COALESCE(parent_user_id, 0), COALESCE(parent_group, ''), name, age, gender FROM children WHERE family_id = ? AND id = ? ORDER BY id`
		args = append(args, sess.ChildID)
	}
	rows, err := s.repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	children := []protocol.Child{}
	for rows.Next() {
		var c protocol.Child
		if err := rows.Scan(&c.ID, &c.FamilyID, &c.ParentUserID, &c.ParentGroup, &c.Name, &c.Age, &c.Gender); err == nil {
			children = append(children, c)
		}
	}
	return children, nil
}

// Dashboard 查询首页聚合数据。
func (s *Service) Dashboard(ctx context.Context, sess protocol.Session, childID int64) (protocol.Dashboard, error) {
	if childID <= 0 || !s.CanAccessChild(ctx, sess, childID) {
		return protocol.Dashboard{}, errors.New("无权访问该孩子档案")
	}
	account, err := s.ensureAccount(ctx, childID)
	if err != nil {
		return protocol.Dashboard{}, err
	}
	records, _ := s.ScoreRecords(ctx, sess, childID, 8)
	tasks, _ := s.TodayTasks(ctx, sess, childID)
	return protocol.Dashboard{Account: account, Records: records, Tasks: tasks}, nil
}

// ScoreRecords 查询积分流水。
func (s *Service) ScoreRecords(ctx context.Context, sess protocol.Session, childID int64, limit int) ([]protocol.ScoreRecord, error) {
	if childID <= 0 || !s.CanAccessChild(ctx, sess, childID) {
		return nil, errors.New("无权访问该孩子档案")
	}
	return s.listScoreRecords(ctx, childID, limit)
}

// CreateScoreRecord 创建手动积分流水。
func (s *Service) CreateScoreRecord(ctx context.Context, sess protocol.Session, req CreateScoreRecordParam) (protocol.Account, error) {
	logger.L().Info("创建积分记录", zap.Int64("child_id", req.ChildID), zap.String("record_type", req.RecordType), zap.Int("score_change", req.ScoreChange), zap.Int64("operator_id", sess.UserID))
	if !CanOperate(sess) {
		return protocol.Account{}, errors.New("只有管理员或家长可以执行该操作")
	}
	if !s.CanAccessChild(ctx, sess, req.ChildID) {
		return protocol.Account{}, errors.New("无权访问该孩子档案")
	}
	account, err := s.ApplyScoreChange(ctx, protocol.ApplyScoreChangeParam{ChildID: req.ChildID, RecordType: req.RecordType, TargetAccount: req.TargetAccount, ItemName: req.ItemName, ScoreChange: req.ScoreChange, Reason: req.Reason, Evidence: req.Evidence, Operator: sess})
	if err != nil {
		return protocol.Account{}, err
	}
	if strings.ToUpper(strings.TrimSpace(req.RecordType)) == "DEDUCT" {
		if err := s.createRepairTaskForDeduct(ctx, req); err != nil {
			logger.L().Warn("扣分后生成修复任务失败", zap.Error(err), zap.Int64("child_id", req.ChildID))
		}
	}
	return account, nil
}

// TodayTasks 查询今日任务。
func (s *Service) TodayTasks(ctx context.Context, sess protocol.Session, childID int64) ([]protocol.TaskInstance, error) {
	if childID <= 0 || !s.CanAccessChild(ctx, sess, childID) {
		return nil, errors.New("无权访问该孩子档案")
	}
	return s.listTodayTasks(ctx, sess, childID)
}

// SubmitTask 提交任务。
func (s *Service) SubmitTask(ctx context.Context, sess protocol.Session, taskID int64, req SubmitTaskParam) error {
	var childID int64
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT child_id FROM task_instances WHERE id = ?`, taskID).Scan(&childID); err != nil {
		return errors.New("任务不存在")
	}
	if !s.CanAccessChild(ctx, sess, childID) {
		return errors.New("无权操作该任务")
	}
	logger.L().Info("提交任务", zap.Int64("task_id", taskID), zap.Int64("child_id", childID), zap.Int64("operator_id", sess.UserID))
	res, err := s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'SUBMITTED', submit_note = ?, submitted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'TODO'`, req.SubmitNote, taskID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("任务状态不可提交")
	}
	return nil
}

// AuditTask 审核任务。
func (s *Service) AuditTask(ctx context.Context, sess protocol.Session, taskID int64, req AuditParam) (protocol.Account, error) {
	logger.L().Info("审核任务", zap.Int64("task_id", taskID), zap.String("result", req.Result), zap.Int64("operator_id", sess.UserID))
	if !CanOperate(sess) {
		return protocol.Account{}, errors.New("只有管理员或家长可以执行该操作")
	}
	var t protocol.TaskInstance
	var questionsRaw string
	err := s.repo.DB.QueryRowContext(ctx, `SELECT id, child_id, COALESCE(template_id, 0), task_name, task_type, category, COALESCE(subject, 'GENERAL'), COALESCE(content, ''), COALESCE(question_type, 'NONE'), COALESCE(answer, ''), COALESCE(questions, '[]'), score_value, target_account, status, submit_note, audit_note, task_date, COALESCE(due_at, '') FROM task_instances WHERE id = ?`, taskID).Scan(&t.ID, &t.ChildID, &t.TemplateID, &t.TaskName, &t.TaskType, &t.Category, &t.Subject, &t.Content, &t.QuestionType, &t.Answer, &questionsRaw, &t.ScoreValue, &t.TargetAccount, &t.Status, &t.SubmitNote, &t.AuditNote, &t.TaskDate, &t.DueAt)
	if err != nil {
		return protocol.Account{}, errors.New("任务不存在")
	}
	t.Questions = decodeTaskQuestions(questionsRaw, t.Subject, t.QuestionType, t.Content, t.Answer)
	if !s.CanAccessChild(ctx, sess, t.ChildID) {
		return protocol.Account{}, errors.New("无权审核该任务")
	}
	if t.Status != "SUBMITTED" {
		return protocol.Account{}, errors.New("任务必须由孩子提交后才能审核")
	}
	result := strings.ToUpper(req.Result)
	if result != "APPROVED" {
		res, err := s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'REJECTED', audit_note = ?, audited_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'SUBMITTED'`, req.AuditNote, taskID)
		if err != nil {
			return protocol.Account{}, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return protocol.Account{}, errors.New("任务已被其他家长处理，请刷新")
		}
		return protocol.Account{}, nil
	}
	res, err := s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'APPROVING', audit_note = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'SUBMITTED'`, req.AuditNote, taskID)
	if err != nil {
		return protocol.Account{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return protocol.Account{}, errors.New("任务已被其他家长处理，请刷新")
	}
	var approvedScore int
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(score_value), 0) FROM task_instances WHERE child_id = ? AND task_date = ? AND status = 'APPROVED'`, t.ChildID, t.TaskDate).Scan(&approvedScore); err != nil {
		_, _ = s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'SUBMITTED', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'APPROVING'`, taskID)
		return protocol.Account{}, err
	}
	if approvedScore+t.ScoreValue > consts.DailyTaskScoreLimit {
		_, _ = s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'SUBMITTED', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'APPROVING'`, taskID)
		return protocol.Account{}, errors.New("当天任务加分不能超过 15 分")
	}
	recordType := "ADD"
	target := "AUTO"
	switch t.TaskType {
	case "REPAIR":
		recordType = "REPAIR"
		target = "BASE"
	case "TEAM":
		recordType = "TEAM"
		target = "TEAM"
	}
	account, err := s.ApplyScoreChange(ctx, protocol.ApplyScoreChangeParam{ChildID: t.ChildID, RecordType: recordType, TargetAccount: target, ItemName: t.TaskName, ScoreChange: t.ScoreValue, Reason: "任务审核通过", Operator: sess})
	if err != nil {
		_, _ = s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'SUBMITTED', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'APPROVING'`, taskID)
		return protocol.Account{}, err
	}
	res, err = s.repo.DB.ExecContext(ctx, `UPDATE task_instances SET status = 'APPROVED', audit_note = ?, audited_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'APPROVING'`, req.AuditNote, taskID)
	if err != nil {
		return protocol.Account{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return protocol.Account{}, errors.New("任务状态异常，请刷新")
	}
	return account, nil
}

// TaskTemplates 查询任务模板列表。
func (s *Service) TaskTemplates(ctx context.Context, sess protocol.Session) ([]protocol.TaskTemplate, error) {
	if !CanOperate(sess) {
		return nil, errors.New("只有管理员或家长可以查看任务配置")
	}
	rows, err := s.repo.DB.QueryContext(ctx, `SELECT id, task_name, task_type, category, COALESCE(subject, 'GENERAL'), COALESCE(content, ''), COALESCE(question_type, 'NONE'), COALESCE(answer, ''), COALESCE(questions, '[]'), score_value, target_account, need_parent_confirm, daily_limit, weekly_limit, enabled, description, COALESCE(due_time, '') FROM task_templates WHERE family_id = ? ORDER BY id DESC`, sess.FamilyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.TaskTemplate{}
	for rows.Next() {
		var it protocol.TaskTemplate
		var questionsRaw string
		if err := rows.Scan(&it.ID, &it.TaskName, &it.TaskType, &it.Category, &it.Subject, &it.Content, &it.QuestionType, &it.Answer, &questionsRaw, &it.ScoreValue, &it.TargetAccount, &it.NeedParentConfirm, &it.DailyLimit, &it.WeeklyLimit, &it.Enabled, &it.Description, &it.DueTime); err == nil {
			it.Questions = decodeTaskQuestions(questionsRaw, it.Subject, it.QuestionType, it.Content, it.Answer)
			items = append(items, it)
		}
	}
	return items, nil
}

// CreateTaskTemplate 创建任务模板。
func (s *Service) CreateTaskTemplate(ctx context.Context, sess protocol.Session, req CreateTaskTemplateParam) (int64, error) {
	if !CanOperate(sess) {
		return 0, errors.New("只有管理员或家长可以执行该操作")
	}
	req.TaskName = strings.TrimSpace(req.TaskName)
	req.TaskType = strings.ToUpper(strings.TrimSpace(req.TaskType))
	req.Category = strings.ToUpper(strings.TrimSpace(req.Category))
	req.Subject = strings.ToUpper(strings.TrimSpace(req.Subject))
	req.QuestionType = strings.ToUpper(strings.TrimSpace(req.QuestionType))
	req.TargetAccount = strings.ToUpper(strings.TrimSpace(req.TargetAccount))
	req.Content = strings.TrimSpace(req.Content)
	req.Answer = strings.TrimSpace(req.Answer)
	req.Questions = normalizeTaskQuestions(req.Questions, req.Subject, req.QuestionType, req.Content, req.Answer)
	questionsJSON := encodeTaskQuestions(req.Questions)
	req.Description = strings.TrimSpace(req.Description)
	req.DueTime = normalizeDueTime(req.DueTime)
	if req.TaskName == "" || req.ScoreValue <= 0 {
		return 0, errors.New("任务名称和分值不能为空")
	}
	if req.TaskType == "" {
		req.TaskType = "DAILY"
	}
	if req.Category == "" {
		req.Category = "ACTION"
	}
	if req.Subject == "" {
		req.Subject = "GENERAL"
	}
	if req.QuestionType == "" {
		req.QuestionType = "NONE"
	}
	if req.TargetAccount == "" {
		req.TargetAccount = "AUTO"
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO task_templates(family_id, task_name, task_type, category, subject, content, question_type, answer, questions, score_value, target_account, description, due_time) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, sess.FamilyID, req.TaskName, req.TaskType, req.Category, req.Subject, req.Content, req.QuestionType, req.Answer, questionsJSON, req.ScoreValue, req.TargetAccount, req.Description, req.DueTime)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("新增任务模板", zap.Int64("template_id", id), zap.String("task_name", req.TaskName), zap.Int64("family_id", sess.FamilyID))
	return id, nil
}

// DeleteTaskTemplate 删除任务模板。
func (s *Service) DeleteTaskTemplate(ctx context.Context, sess protocol.Session, id int64) error {
	if !CanOperate(sess) {
		return errors.New("只有管理员或家长可以执行该操作")
	}
	_, err := s.repo.DB.ExecContext(ctx, `DELETE FROM task_templates WHERE id = ? AND family_id = ?`, id, sess.FamilyID)
	if err == nil {
		logger.L().Info("删除任务模板", zap.Int64("template_id", id), zap.Int64("family_id", sess.FamilyID))
	}
	return err
}

// Rewards 查询奖励列表。
func (s *Service) Rewards(ctx context.Context, sess protocol.Session) ([]protocol.Reward, error) {
	rows, err := s.repo.DB.QueryContext(ctx, `SELECT id, reward_name, reward_type, cost_score, cost_star, weekly_limit, monthly_limit, health_risk, need_parent_confirm, enabled, description FROM rewards WHERE family_id = ? ORDER BY cost_star, cost_score`, sess.FamilyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rewards := []protocol.Reward{}
	for rows.Next() {
		var rw protocol.Reward
		if err := rows.Scan(&rw.ID, &rw.RewardName, &rw.RewardType, &rw.CostScore, &rw.CostStar, &rw.WeeklyLimit, &rw.MonthlyLimit, &rw.HealthRisk, &rw.NeedParentConfirm, &rw.Enabled, &rw.Description); err == nil {
			rewards = append(rewards, rw)
		}
	}
	return rewards, nil
}

// CreateReward 创建奖励项。
func (s *Service) CreateReward(ctx context.Context, sess protocol.Session, req CreateRewardParam) (int64, error) {
	if !CanOperate(sess) {
		return 0, errors.New("只有管理员或家长可以执行该操作")
	}
	req.RewardName = strings.TrimSpace(req.RewardName)
	req.RewardType = strings.ToUpper(strings.TrimSpace(req.RewardType))
	req.HealthRisk = strings.ToUpper(strings.TrimSpace(req.HealthRisk))
	if req.RewardName == "" || (req.CostScore <= 0 && req.CostStar <= 0) {
		return 0, errors.New("奖励名称和兑换分值不能为空")
	}
	if req.RewardType == "" {
		req.RewardType = "SNACK"
	}
	if req.HealthRisk == "" {
		req.HealthRisk = "NONE"
	}
	if req.WeeklyLimit <= 0 {
		req.WeeklyLimit = 1
	}
	if req.MonthlyLimit <= 0 {
		req.MonthlyLimit = 4
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO rewards(family_id, reward_name, reward_type, cost_score, cost_star, weekly_limit, monthly_limit, health_risk, description) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, sess.FamilyID, req.RewardName, req.RewardType, req.CostScore, req.CostStar, req.WeeklyLimit, req.MonthlyLimit, req.HealthRisk, req.Description)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("新增奖励项", zap.Int64("reward_id", id), zap.String("reward_name", req.RewardName), zap.Int64("family_id", sess.FamilyID))
	return id, nil
}

// DeleteReward 删除奖励项。
func (s *Service) DeleteReward(ctx context.Context, sess protocol.Session, id int64) error {
	if !CanOperate(sess) {
		return errors.New("只有管理员或家长可以执行该操作")
	}
	_, err := s.repo.DB.ExecContext(ctx, `DELETE FROM rewards WHERE id = ? AND family_id = ?`, id, sess.FamilyID)
	if err == nil {
		logger.L().Info("删除奖励项", zap.Int64("reward_id", id), zap.Int64("family_id", sess.FamilyID))
	}
	return err
}

// CreateExchangeOrder 创建兑换申请。
func (s *Service) CreateExchangeOrder(ctx context.Context, sess protocol.Session, req CreateExchangeOrderParam) (int64, error) {
	logger.L().Info("创建兑换申请", zap.Int64("child_id", req.ChildID), zap.Int64("reward_id", req.RewardID), zap.Int64("operator_id", sess.UserID))
	if !s.CanAccessChild(ctx, sess, req.ChildID) {
		return 0, errors.New("无权访问该孩子档案")
	}
	var rw protocol.Reward
	err := s.repo.DB.QueryRowContext(ctx, `SELECT id, reward_name, reward_type, cost_score, cost_star, weekly_limit, monthly_limit, health_risk, need_parent_confirm, enabled, description FROM rewards WHERE id = ? AND family_id = ?`, req.RewardID, sess.FamilyID).Scan(&rw.ID, &rw.RewardName, &rw.RewardType, &rw.CostScore, &rw.CostStar, &rw.WeeklyLimit, &rw.MonthlyLimit, &rw.HealthRisk, &rw.NeedParentConfirm, &rw.Enabled, &rw.Description)
	if err != nil || !rw.Enabled || rw.HealthRisk == "HIGH" {
		return 0, errors.New("兑换项不可用")
	}
	account, err := s.ensureAccount(ctx, req.ChildID)
	if err != nil {
		return 0, err
	}
	if account.BaseScore < 80 {
		return 0, errors.New("基准分低于 80，暂停全部兑换")
	}
	if account.BaseScore < 90 && rw.CostScore >= 15 {
		return 0, errors.New("基准分低于 90，暂停高价值兑换")
	}
	if account.BonusScore < rw.CostScore || account.StarCount < rw.CostStar {
		return 0, errors.New("可兑换积分或星星不足")
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO exchange_orders(child_id, reward_id, cost_score, cost_star, apply_note) VALUES(?, ?, ?, ?, ?)`, req.ChildID, req.RewardID, rw.CostScore, rw.CostStar, req.Note)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ExchangeOrders 查询兑换申请。
func (s *Service) ExchangeOrders(ctx context.Context, sess protocol.Session) ([]protocol.ExchangeOrder, error) {
	query := `SELECT o.id, o.child_id, o.reward_id, r.reward_name, o.cost_score, o.cost_star, o.status, o.apply_note, o.audit_note, o.applied_at FROM exchange_orders o JOIN rewards r ON r.id = o.reward_id JOIN children c ON c.id = o.child_id WHERE c.family_id = ? ORDER BY o.id DESC LIMIT 100`
	args := []any{sess.FamilyID}
	switch sess.Role {
	case consts.RoleChild:
		query = `SELECT o.id, o.child_id, o.reward_id, r.reward_name, o.cost_score, o.cost_star, o.status, o.apply_note, o.audit_note, o.applied_at FROM exchange_orders o JOIN rewards r ON r.id = o.reward_id JOIN children c ON c.id = o.child_id WHERE c.family_id = ? AND o.child_id = ? ORDER BY o.id DESC LIMIT 100`
		args = append(args, sess.ChildID)
	case consts.RoleParent:
		parentGroup := s.parentGroupOf(ctx, sess.FamilyID, sess.UserID)
		query = `SELECT o.id, o.child_id, o.reward_id, r.reward_name, o.cost_score, o.cost_star, o.status, o.apply_note, o.audit_note, o.applied_at FROM exchange_orders o JOIN rewards r ON r.id = o.reward_id JOIN children c ON c.id = o.child_id WHERE c.family_id = ? AND (c.parent_user_id = ? OR (? <> '' AND c.parent_group = ?)) ORDER BY o.id DESC LIMIT 100`
		args = append(args, sess.UserID, parentGroup, parentGroup)
	}
	rows, err := s.repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.ExchangeOrder{}
	for rows.Next() {
		var it protocol.ExchangeOrder
		if err := rows.Scan(&it.ID, &it.ChildID, &it.RewardID, &it.RewardName, &it.CostScore, &it.CostStar, &it.Status, &it.ApplyNote, &it.AuditNote, &it.AppliedAt); err == nil {
			items = append(items, it)
		}
	}
	return items, nil
}

// AuditExchangeOrder 审核兑换申请。
func (s *Service) AuditExchangeOrder(ctx context.Context, sess protocol.Session, orderID int64, req AuditParam) (protocol.Account, error) {
	logger.L().Info("审核兑换申请", zap.Int64("order_id", orderID), zap.String("result", req.Result), zap.Int64("operator_id", sess.UserID))
	if !CanOperate(sess) {
		return protocol.Account{}, errors.New("只有管理员或家长可以执行该操作")
	}
	var childID, rewardID int64
	var costScore, costStar int
	var status string
	err := s.repo.DB.QueryRowContext(ctx, `SELECT child_id, reward_id, cost_score, cost_star, status FROM exchange_orders WHERE id = ?`, orderID).Scan(&childID, &rewardID, &costScore, &costStar, &status)
	if err != nil {
		return protocol.Account{}, errors.New("兑换申请不存在")
	}
	if !s.CanAccessChild(ctx, sess, childID) {
		return protocol.Account{}, errors.New("无权审核该兑换申请")
	}
	if status != "PENDING" {
		return protocol.Account{}, errors.New("兑换申请状态不可审核")
	}
	if strings.ToUpper(req.Result) != "APPROVED" {
		_, _ = s.repo.DB.ExecContext(ctx, `UPDATE exchange_orders SET status = 'REJECTED', audit_note = ?, audited_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'PENDING'`, req.AuditNote, orderID)
		return protocol.Account{}, nil
	}
	if err := s.validateExchangeApproval(ctx, sess.FamilyID, childID, rewardID, costScore, costStar); err != nil {
		return protocol.Account{}, err
	}
	return s.applyExchange(ctx, childID, orderID, rewardID, costScore, costStar, sess, req.AuditNote)
}

func (s *Service) validateExchangeApproval(ctx context.Context, familyID, childID, rewardID int64, costScore, costStar int) error {
	var rw protocol.Reward
	err := s.repo.DB.QueryRowContext(ctx, `SELECT id, reward_name, reward_type, cost_score, cost_star, weekly_limit, monthly_limit, health_risk, need_parent_confirm, enabled, description FROM rewards WHERE id = ? AND family_id = ?`, rewardID, familyID).Scan(&rw.ID, &rw.RewardName, &rw.RewardType, &rw.CostScore, &rw.CostStar, &rw.WeeklyLimit, &rw.MonthlyLimit, &rw.HealthRisk, &rw.NeedParentConfirm, &rw.Enabled, &rw.Description)
	if err != nil || !rw.Enabled || rw.HealthRisk == "HIGH" {
		return errors.New("兑换项不可用")
	}
	account, err := s.ensureAccount(ctx, childID)
	if err != nil {
		return err
	}
	if account.BaseScore < 80 {
		return errors.New("基准分低于 80，暂停全部兑换")
	}
	if account.BaseScore < 90 && costScore >= 15 {
		return errors.New("基准分低于 90，暂停高价值兑换")
	}
	if account.BonusScore < costScore || account.StarCount < costStar {
		return errors.New("可兑换积分或星星不足")
	}
	if rw.WeeklyLimit > 0 {
		var weekCount int
		if err := s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM exchange_orders WHERE child_id = ? AND reward_id = ? AND status = 'APPROVED' AND completed_at >= DATETIME('now', 'localtime', '-7 days')`, childID, rewardID).Scan(&weekCount); err != nil {
			return err
		}
		if weekCount >= rw.WeeklyLimit {
			return errors.New("该奖励本周兑换次数已达上限")
		}
	}
	if rw.MonthlyLimit > 0 {
		var monthCount int
		if err := s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM exchange_orders WHERE child_id = ? AND reward_id = ? AND status = 'APPROVED' AND STRFTIME('%Y-%m', completed_at) = STRFTIME('%Y-%m', 'now', 'localtime')`, childID, rewardID).Scan(&monthCount); err != nil {
			return err
		}
		if monthCount >= rw.MonthlyLimit {
			return errors.New("该奖励本月兑换次数已达上限")
		}
	}
	return nil
}

// ApplyScoreChange 执行账户分值变更并写入流水。
func (s *Service) ApplyScoreChange(ctx context.Context, p protocol.ApplyScoreChangeParam) (protocol.Account, error) {
	logger.L().Info("执行账户分值变更", zap.Int64("child_id", p.ChildID), zap.String("record_type", p.RecordType), zap.String("target_account", p.TargetAccount), zap.Int("score_change", p.ScoreChange))
	if strings.TrimSpace(p.ItemName) == "" {
		return protocol.Account{}, errors.New("项目名称不能为空")
	}
	if p.ScoreChange == 0 {
		return protocol.Account{}, errors.New("分值不能为 0")
	}
	p.RecordType = strings.ToUpper(strings.TrimSpace(p.RecordType))
	p.TargetAccount = strings.ToUpper(strings.TrimSpace(p.TargetAccount))
	if p.TargetAccount == "" {
		p.TargetAccount = "AUTO"
	}
	if p.RecordType == "" {
		p.RecordType = "ADD"
	}

	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return protocol.Account{}, err
	}
	defer tx.Rollback()
	account, err := ensureAccountTx(ctx, tx, p.ChildID)
	if err != nil {
		return protocol.Account{}, err
	}
	var before int
	var after int
	change := utils.Abs(p.ScoreChange)
	signedChange := change
	var targetForRecord string
	switch p.RecordType {
	case "DEDUCT":
		before = account.BaseScore
		account.BaseScore -= change
		if account.BaseScore < 0 {
			account.BaseScore = 0
		}
		after = account.BaseScore
		signedChange = -change
		targetForRecord = "BASE"
	case "REPAIR":
		before = account.BaseScore
		account.BaseScore += change
		if account.BaseScore > 100 {
			account.BaseScore = 100
		}
		after = account.BaseScore
		targetForRecord = "BASE"
	case "TEAM":
		before = account.TeamScore
		account.TeamScore += change
		after = account.TeamScore
		targetForRecord = "TEAM"
	case "STAR":
		before = account.StarCount
		account.StarCount += change
		if account.StarCount > consts.StarLimit {
			account.StarCount = 20
		}
		after = account.StarCount
		targetForRecord = "STAR"
	case "ADD":
		if account.BaseScore < 100 {
			before = account.BaseScore
			restore := utils.Min(change, 100-account.BaseScore)
			account.BaseScore += restore
			remain := change - restore
			if remain > 0 {
				account.BonusScore += remain
			}
			after = account.BaseScore
			targetForRecord = "AUTO"
		} else {
			before = account.BonusScore
			account.BonusScore += change
			after = account.BonusScore
			targetForRecord = "BONUS"
		}
	default:
		return protocol.Account{}, errors.New("不支持的明细类型")
	}
	account.StatusLevel = calcStatus(account.BaseScore)
	if _, err := tx.ExecContext(ctx, `INSERT INTO score_records(child_id, account_id, record_type, target_account, item_name, score_change, before_value, after_value, operator_role, operator_id, reason, evidence) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.ChildID, account.ID, p.RecordType, targetForRecord, p.ItemName, signedChange, before, after, p.Operator.Role, p.Operator.UserID, p.Reason, p.Evidence); err != nil {
		return protocol.Account{}, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE score_accounts SET base_score = ?, bonus_score = ?, star_count = ?, team_score = ?, status_level = ?, version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND version = ?`, account.BaseScore, account.BonusScore, account.StarCount, account.TeamScore, account.StatusLevel, account.ID, account.Version)
	if err != nil {
		return protocol.Account{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return protocol.Account{}, errors.New("账户版本冲突，请刷新后重试")
	}
	account.Version++
	if err := tx.Commit(); err != nil {
		logger.L().Error("账户分值变更事务提交失败", zap.Error(err), zap.Int64("child_id", p.ChildID))
		return protocol.Account{}, err
	}
	logger.L().Info("账户分值变更完成", zap.Int64("child_id", p.ChildID), zap.String("record_type", p.RecordType), zap.String("status_level", account.StatusLevel), zap.Int("base_score", account.BaseScore), zap.Int("bonus_score", account.BonusScore))
	return account, nil
}

func (s *Service) applyExchange(ctx context.Context, childID, orderID, rewardID int64, costScore, costStar int, sess protocol.Session, note string) (protocol.Account, error) {
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return protocol.Account{}, err
	}
	defer tx.Rollback()
	account, err := ensureAccountTx(ctx, tx, childID)
	if err != nil {
		return protocol.Account{}, err
	}
	if account.BonusScore < costScore || account.StarCount < costStar {
		return protocol.Account{}, errors.New("可兑换积分或星星不足")
	}
	beforeBonus := account.BonusScore
	account.BonusScore -= costScore
	account.StarCount -= costStar
	if account.StarCount < 0 {
		account.StarCount = 0
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO score_records(child_id, account_id, record_type, target_account, item_name, score_change, before_value, after_value, operator_role, operator_id, reason) VALUES(?, ?, 'EXCHANGE', 'BONUS', ?, ?, ?, ?, ?, ?, ?)`, childID, account.ID, fmt.Sprintf("兑换奖励 #%d", rewardID), -costScore, beforeBonus, account.BonusScore, sess.Role, sess.UserID, note); err != nil {
		return protocol.Account{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE score_accounts SET bonus_score = ?, star_count = ?, last_exchange_date = DATE('now'), version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND version = ?`, account.BonusScore, account.StarCount, account.ID, account.Version); err != nil {
		return protocol.Account{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE exchange_orders SET status = 'APPROVED', audit_note = ?, audited_at = CURRENT_TIMESTAMP, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, note, orderID); err != nil {
		return protocol.Account{}, err
	}
	account.Version++
	if err := tx.Commit(); err != nil {
		return protocol.Account{}, err
	}
	return account, nil
}

func (s *Service) ensureAccount(ctx context.Context, childID int64) (protocol.Account, error) {
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return protocol.Account{}, err
	}
	defer tx.Rollback()
	account, err := ensureAccountTx(ctx, tx, childID)
	if err != nil {
		return protocol.Account{}, err
	}
	return account, tx.Commit()
}

func ensureAccountTx(ctx context.Context, tx *sql.Tx, childID int64) (protocol.Account, error) {
	month := utils.CurrentMonth()
	var a protocol.Account
	err := tx.QueryRowContext(ctx, `SELECT id, child_id, base_score, bonus_score, star_count, team_score, status_level, current_month, COALESCE(last_exchange_date, ''), appeal_count_this_week, version FROM score_accounts WHERE child_id = ? AND current_month = ?`, childID, month).Scan(&a.ID, &a.ChildID, &a.BaseScore, &a.BonusScore, &a.StarCount, &a.TeamScore, &a.StatusLevel, &a.CurrentMonth, &a.LastExchangeDate, &a.AppealCountThisWeek, &a.Version)
	if err == nil {
		return a, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return protocol.Account{}, err
	}
	star := 0
	_ = tx.QueryRowContext(ctx, `SELECT star_count FROM score_accounts WHERE child_id = ? ORDER BY current_month DESC LIMIT 1`, childID).Scan(&star)
	res, err := tx.ExecContext(ctx, `INSERT INTO score_accounts(child_id, base_score, bonus_score, star_count, team_score, status_level, current_month) VALUES(?, 100, 0, ?, 0, 'GREEN', ?)`, childID, star, month)
	if err != nil {
		return protocol.Account{}, err
	}
	id, _ := res.LastInsertId()
	return protocol.Account{ID: id, ChildID: childID, BaseScore: 100, StarCount: star, StatusLevel: "GREEN", CurrentMonth: month}, nil
}

func (s *Service) listScoreRecords(ctx context.Context, childID int64, limit int) ([]protocol.ScoreRecord, error) {
	rows, err := s.repo.DB.QueryContext(ctx, `SELECT id, child_id, record_type, target_account, item_name, score_change, before_value, after_value, operator_role, reason, evidence, confirm_status, occurred_at FROM score_records WHERE child_id = ? ORDER BY occurred_at DESC, id DESC LIMIT ?`, childID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.ScoreRecord{}
	for rows.Next() {
		var it protocol.ScoreRecord
		if err := rows.Scan(&it.ID, &it.ChildID, &it.RecordType, &it.TargetAccount, &it.ItemName, &it.ScoreChange, &it.BeforeValue, &it.AfterValue, &it.OperatorRole, &it.Reason, &it.Evidence, &it.ConfirmStatus, &it.OccurredAt); err == nil {
			items = append(items, it)
		}
	}
	return items, nil
}

func (s *Service) listTodayTasks(ctx context.Context, sess protocol.Session, childID int64) ([]protocol.TaskInstance, error) {
	if err := s.ensureTodayTasks(ctx, sess.FamilyID, childID); err != nil {
		return nil, err
	}
	rows, err := s.repo.DB.QueryContext(ctx, `SELECT id, child_id, COALESCE(template_id, 0), task_name, task_type, category, COALESCE(subject, 'GENERAL'), COALESCE(content, ''), COALESCE(question_type, 'NONE'), COALESCE(answer, ''), COALESCE(questions, '[]'), score_value, target_account, status, submit_note, audit_note, task_date, COALESCE(due_at, '') FROM task_instances WHERE child_id = ? AND task_date = DATE('now', 'localtime') ORDER BY CASE WHEN due_at IS NULL OR due_at = '' THEN 1 ELSE 0 END, due_at, id`, childID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.TaskInstance{}
	for rows.Next() {
		var t protocol.TaskInstance
		var questionsRaw string
		if err := rows.Scan(&t.ID, &t.ChildID, &t.TemplateID, &t.TaskName, &t.TaskType, &t.Category, &t.Subject, &t.Content, &t.QuestionType, &t.Answer, &questionsRaw, &t.ScoreValue, &t.TargetAccount, &t.Status, &t.SubmitNote, &t.AuditNote, &t.TaskDate, &t.DueAt); err == nil {
			t.Questions = decodeTaskQuestions(questionsRaw, t.Subject, t.QuestionType, t.Content, t.Answer)
			items = append(items, t)
		}
	}
	return items, nil
}

func (s *Service) ensureTodayTasks(ctx context.Context, familyID, childID int64) error {
	rows, err := s.repo.DB.QueryContext(ctx, `SELECT id, task_name, task_type, category, COALESCE(subject, 'GENERAL'), COALESCE(content, ''), COALESCE(question_type, 'NONE'), COALESCE(answer, ''), COALESCE(questions, '[]'), score_value, target_account, COALESCE(due_time, '') FROM task_templates WHERE family_id = ? AND task_type IN ('DAILY', 'TEAM') AND enabled = 1`, familyID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var templateID int64
		var name, typ, cat, subject, content, questionType, answer, questionsRaw, target, dueTime string
		var score int
		if err := rows.Scan(&templateID, &name, &typ, &cat, &subject, &content, &questionType, &answer, &questionsRaw, &score, &target, &dueTime); err != nil {
			continue
		}
		var exists int
		_ = s.repo.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM task_instances WHERE child_id = ? AND template_id = ? AND task_date = DATE('now', 'localtime')`, childID, templateID).Scan(&exists)
		if exists == 0 {
			_, err = s.repo.DB.ExecContext(ctx, `INSERT INTO task_instances(child_id, template_id, task_name, task_type, category, subject, content, question_type, answer, questions, score_value, target_account, task_date, due_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, DATE('now', 'localtime'), CASE WHEN ? <> '' THEN DATETIME(DATE('now', 'localtime') || ' ' || ?) ELSE NULL END)`, childID, templateID, name, typ, cat, subject, content, questionType, answer, questionsRaw, score, target, dueTime, dueTime)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CanAccessChild 判断当前会话是否可访问孩子档案。
func (s *Service) CanAccessChild(ctx context.Context, sess protocol.Session, childID int64) bool {
	if sess.Role == consts.RoleChild && sess.ChildID != childID {
		return false
	}
	query := `SELECT COUNT(*) FROM children WHERE id = ? AND family_id = ?`
	args := []any{childID, sess.FamilyID}
	if sess.Role == consts.RoleParent {
		parentGroup := s.parentGroupOf(ctx, sess.FamilyID, sess.UserID)
		query = `SELECT COUNT(*) FROM children WHERE id = ? AND family_id = ? AND (parent_user_id = ? OR (? <> '' AND parent_group = ?))`
		args = append(args, sess.UserID, parentGroup, parentGroup)
	}
	var count int
	_ = s.repo.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	return count > 0
}

func (s *Service) parentGroupOf(ctx context.Context, familyID, userID int64) string {
	var group string
	_ = s.repo.DB.QueryRowContext(ctx, `SELECT COALESCE(parent_group, '') FROM users WHERE id = ? AND family_id = ? AND role = 'PARENT' AND enabled = 1`, userID, familyID).Scan(&group)
	return group
}

func normalizeParentTitle(title string) string {
	title = strings.TrimSpace(title)
	allowed := map[string]struct{}{"爸爸": {}, "妈妈": {}, "爷爷": {}, "奶奶": {}, "姥姥": {}, "姥爷": {}}
	if _, ok := allowed[title]; ok {
		return title
	}
	return "爸爸"
}

func normalizeParentGroup(group string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return "默认监护组"
	}
	if len([]rune(group)) > 32 {
		return string([]rune(group)[:32])
	}
	return group
}
