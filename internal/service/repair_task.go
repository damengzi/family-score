package service

import (
	"context"
	"strings"

	"family-score/internal/protocol"
	"family-score/pkg/utils"
)

func (s *Service) createRepairTaskForDeduct(ctx context.Context, req protocol.CreateScoreRecordParam) error {
	name := strings.TrimSpace(req.ItemName)
	if name == "" {
		name = "扣分修复"
	}
	score := utils.Abs(req.ScoreChange)
	if score <= 0 {
		score = 1
	}
	content := strings.TrimSpace(req.Reason)
	if content == "" {
		content = "请说明这次发生了什么、影响了谁、下次准备怎么做。"
	}
	_, err := s.repo.DB.ExecContext(ctx, `INSERT INTO task_instances(child_id, template_id, task_name, task_type, category, subject, content, question_type, answer, score_value, target_account, task_date, due_at) VALUES(?, NULL, ?, 'REPAIR', 'EMOTION', 'MORAL', ?, 'SHORT_ANSWER', ?, ?, 'BASE', DATE('now', 'localtime'), DATETIME(DATE('now', 'localtime') || ' 23:59'))`, req.ChildID, "修复任务："+name, content, "能说出原因、影响和下次做法，并完成补救行动", score)
	return err
}
