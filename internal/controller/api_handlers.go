package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"family-score/internal/logger"
	"family-score/internal/protocol"
	"family-score/internal/service"

	"go.uber.org/zap"
)

// Controller 表示控制层，负责 HTTP 入参出参和状态码处理。
type Controller struct {
	svc *service.Service
}

// New 创建控制层实例。
func New(svc *service.Service) *Controller { return &Controller{svc: svc} }

// Service 返回控制层持有的服务层。
func (c *Controller) Service() *service.Service { return c.svc }

// SystemStatus 查询系统状态。
func (c *Controller) SystemStatus(w http.ResponseWriter, r *http.Request) {
	status, err := c.svc.SystemStatus(r.Context())
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

// SetupInit 初始化系统。
func (c *Controller) SetupInit(w http.ResponseWriter, r *http.Request) {
	var req service.SetupInitParam
	if !readJSON(w, r, &req) {
		return
	}
	if err := c.svc.SetupInit(r.Context(), req); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Register 自主注册普通用户。
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	var req service.SelfRegisterParam
	if !readJSON(w, r, &req) {
		return
	}
	id, err := c.svc.SelfRegister(r.Context(), req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "role": "NORMAL"})
}

// Login 登录。
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req struct{ LoginName, Password string }
	if !readJSON(w, r, &req) {
		return
	}
	token, sess, err := c.svc.Login(r.Context(), req.LoginName, req.Password)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "fs_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Expires: sess.ExpiresAt})
	writeJSON(w, http.StatusOK, map[string]any{"user": sess})
}

// PasswordCaptcha 获取忘记密码图片验证码。
func (c *Controller) PasswordCaptcha(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, c.svc.PasswordCaptcha())
}

// ResetPassword 通过图片验证码重置密码。
func (c *Controller) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req protocol.ResetPasswordParam
	if !readJSON(w, r, &req) {
		return
	}
	if err := c.svc.ResetPassword(r.Context(), req); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// AuthMe 查询当前登录用户。
func (c *Controller) AuthMe(w http.ResponseWriter, _ *http.Request, sess protocol.Session) {
	writeJSON(w, http.StatusOK, map[string]any{"user": sess})
}

// Profile 查询当前登录用户个人主页信息。
func (c *Controller) Profile(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	profile, err := c.svc.Profile(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": profile})
}

// ChangeMyPassword 修改当前登录用户自己的密码。
func (c *Controller) ChangeMyPassword(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req protocol.ChangePasswordParam
	if !readJSON(w, r, &req) {
		return
	}
	if err := c.svc.ChangeMyPassword(r.Context(), sess, req); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Users 查询用户列表。
func (c *Controller) Users(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	users, err := c.svc.Users(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// CreateUser 注册用户。
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req service.CreateUserParam
	if !readJSON(w, r, &req) {
		return
	}
	id, err := c.svc.CreateUser(r.Context(), sess, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

// UpdateUser 修改用户。
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request, sess protocol.Session, id int64) {
	var req protocol.UpdateUserParam
	if !readJSON(w, r, &req) {
		return
	}
	user, err := c.svc.UpdateUser(r.Context(), sess, id, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// DeleteUser 注销用户。
func (c *Controller) DeleteUser(w http.ResponseWriter, r *http.Request, sess protocol.Session, id int64) {
	if err := c.svc.DeleteUser(r.Context(), sess, id); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Logout 退出登录。
func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("fs_session"); err == nil {
		c.svc.Logout(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "fs_session", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// RequireSession 从 Cookie 中获取登录会话。
func (c *Controller) RequireSession(w http.ResponseWriter, r *http.Request) (protocol.Session, bool) {
	cookie, err := r.Cookie("fs_session")
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "请先登录")
		return protocol.Session{}, false
	}
	sess, ok := c.svc.Session(cookie.Value)
	if !ok {
		errorJSON(w, http.StatusUnauthorized, "登录已过期")
		return protocol.Session{}, false
	}
	return sess, true
}

// Children 查询孩子列表。
func (c *Controller) Children(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	children, err := c.svc.Children(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"children": children})
}

// CreateChild 创建孩子档案。
func (c *Controller) CreateChild(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req service.CreateChildParam
	if !readJSON(w, r, &req) {
		return
	}
	id, err := c.svc.CreateChild(r.Context(), sess, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

// UpdateChild 修改孩子档案。
func (c *Controller) UpdateChild(w http.ResponseWriter, r *http.Request, sess protocol.Session, id int64) {
	var req protocol.UpdateChildParam
	if !readJSON(w, r, &req) {
		return
	}
	child, err := c.svc.UpdateChild(r.Context(), sess, id, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"child": child})
}

// DeleteChild 删除孩子档案。
func (c *Controller) DeleteChild(w http.ResponseWriter, r *http.Request, sess protocol.Session, id int64) {
	if err := c.svc.DeleteChild(r.Context(), sess, id); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Dashboard 查询首页看板。
func (c *Controller) Dashboard(w http.ResponseWriter, r *http.Request, sess protocol.Session, childID int64) {
	dashboard, err := c.svc.Dashboard(r.Context(), sess, childID)
	if err != nil {
		errorJSON(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

// ScoreRecords 查询积分流水。
func (c *Controller) ScoreRecords(w http.ResponseWriter, r *http.Request, sess protocol.Session, childID int64) {
	records, err := c.svc.ScoreRecords(r.Context(), sess, childID, 100)
	if err != nil {
		errorJSON(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": records})
}

// CreateScoreRecord 创建手动记分。
func (c *Controller) CreateScoreRecord(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req service.CreateScoreRecordParam
	if !readJSON(w, r, &req) {
		return
	}
	account, err := c.svc.CreateScoreRecord(r.Context(), sess, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": account})
}

// TodayTasks 查询今日任务。
func (c *Controller) TodayTasks(w http.ResponseWriter, r *http.Request, sess protocol.Session, childID int64) {
	tasks, err := c.svc.TodayTasks(r.Context(), sess, childID)
	if err != nil {
		errorJSON(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

// SubmitTask 提交任务。
func (c *Controller) SubmitTask(w http.ResponseWriter, r *http.Request, sess protocol.Session, taskID int64) {
	var req service.SubmitTaskParam
	if !readJSON(w, r, &req) {
		return
	}
	if err := c.svc.SubmitTask(r.Context(), sess, taskID, req); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// AuditTask 审核任务。
func (c *Controller) AuditTask(w http.ResponseWriter, r *http.Request, sess protocol.Session, taskID int64) {
	var req service.AuditParam
	if !readJSON(w, r, &req) {
		return
	}
	account, err := c.svc.AuditTask(r.Context(), sess, taskID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": account})
}

// TaskTemplates 查询任务模板列表。
func (c *Controller) TaskTemplates(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	items, err := c.svc.TaskTemplates(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"taskTemplates": items})
}

// CreateTaskTemplate 创建任务模板。
func (c *Controller) CreateTaskTemplate(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req service.CreateTaskTemplateParam
	if !readJSON(w, r, &req) {
		return
	}
	id, err := c.svc.CreateTaskTemplate(r.Context(), sess, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

// DeleteTaskTemplate 删除任务模板。
func (c *Controller) DeleteTaskTemplate(w http.ResponseWriter, r *http.Request, sess protocol.Session, id int64) {
	if err := c.svc.DeleteTaskTemplate(r.Context(), sess, id); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Rewards 查询奖励列表。
func (c *Controller) Rewards(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	rewards, err := c.svc.Rewards(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rewards": rewards})
}

// CreateReward 创建奖励项。
func (c *Controller) CreateReward(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req service.CreateRewardParam
	if !readJSON(w, r, &req) {
		return
	}
	id, err := c.svc.CreateReward(r.Context(), sess, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

// DeleteReward 删除奖励项。
func (c *Controller) DeleteReward(w http.ResponseWriter, r *http.Request, sess protocol.Session, id int64) {
	if err := c.svc.DeleteReward(r.Context(), sess, id); err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// CreateExchangeOrder 创建兑换申请。
func (c *Controller) CreateExchangeOrder(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	var req service.CreateExchangeOrderParam
	if !readJSON(w, r, &req) {
		return
	}
	orderID, err := c.svc.CreateExchangeOrder(r.Context(), sess, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orderId": orderID, "status": "PENDING"})
}

// ExchangeOrders 查询兑换申请。
func (c *Controller) ExchangeOrders(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	orders, err := c.svc.ExchangeOrders(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"exchangeOrders": orders})
}

// AuditExchangeOrder 审核兑换申请。
func (c *Controller) AuditExchangeOrder(w http.ResponseWriter, r *http.Request, sess protocol.Session, orderID int64) {
	var req service.AuditParam
	if !readJSON(w, r, &req) {
		return
	}
	account, err := c.svc.AuditExchangeOrder(r.Context(), sess, orderID, req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": account})
}

// Backup 创建本地备份。
func (c *Controller) Backup(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	filePath, fileSize, err := c.svc.Backup(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"filePath": filePath, "fileSize": fileSize})
}

// Backups 查询备份记录。
func (c *Controller) Backups(w http.ResponseWriter, r *http.Request, sess protocol.Session) {
	backups, err := c.svc.Backups(r.Context(), sess)
	if err != nil {
		errorJSON(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"backups": backups})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errorJSON(w http.ResponseWriter, status int, msg string) {
	logger.L().Error("接口处理失败", zap.Int("status", status), zap.String("message", msg))
	writeJSON(w, status, map[string]any{"error": msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "请求读取失败")
		return false
	}
	if err := json.Unmarshal(body, v); err != nil {
		errorJSON(w, http.StatusBadRequest, "JSON 格式错误："+err.Error())
		return false
	}
	return true
}

// SessionExpireDuration 返回默认登录态有效期。
func SessionExpireDuration() time.Duration { return 12 * time.Hour }
