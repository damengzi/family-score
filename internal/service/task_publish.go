package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"family-score/internal/protocol"
)

// PublishTask 给指定孩子发布一次性任务。
func (s *Service) PublishTask(ctx context.Context, sess protocol.Session, req protocol.PublishTaskParam) (int64, error) {
	if !CanOperate(sess) {
		return 0, errors.New("只有管理员或家长可以发布任务")
	}
	if req.ChildID <= 0 || !s.CanAccessChild(ctx, sess, req.ChildID) {
		return 0, errors.New("无权访问该孩子档案")
	}
	req.TaskName = strings.TrimSpace(req.TaskName)
	req.TaskType = strings.ToUpper(strings.TrimSpace(req.TaskType))
	req.Category = strings.ToUpper(strings.TrimSpace(req.Category))
	req.Subject = strings.ToUpper(strings.TrimSpace(req.Subject))
	req.QuestionType = strings.ToUpper(strings.TrimSpace(req.QuestionType))
	req.TargetAccount = strings.ToUpper(strings.TrimSpace(req.TargetAccount))
	req.Content = strings.TrimSpace(req.Content)
	req.Answer = strings.TrimSpace(req.Answer)
	req.TaskDate = strings.TrimSpace(req.TaskDate)
	req.DueAt = strings.ReplaceAll(strings.TrimSpace(req.DueAt), "T", " ")
	req.DueTime = normalizeDueTime(req.DueTime)
	if req.TaskName == "" || req.ScoreValue <= 0 {
		return 0, errors.New("任务名称和分值不能为空")
	}
	if req.TaskType == "" {
		req.TaskType = "TEMP"
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
	taskDate := req.TaskDate
	if taskDate == "" {
		taskDate = time.Now().Format("2006-01-02")
	}
	var dueAt any
	if req.DueAt != "" {
		dueAt = req.DueAt
	} else if req.DueTime != "" {
		dueAt = taskDate + " " + req.DueTime
	}
	args := []any{req.ChildID, req.TaskName, req.TaskType, req.Category, req.Subject, req.Content, req.QuestionType, req.Answer, req.ScoreValue, req.TargetAccount, taskDate, dueAt}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO task_instances(child_id, template_id, task_name, task_type, category, subject, content, question_type, answer, score_value, target_account, task_date, due_at) VALUES(?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
