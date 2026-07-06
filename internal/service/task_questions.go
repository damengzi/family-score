package service

import (
	"encoding/json"
	"strings"

	"family-score/internal/protocol"
)

func normalizeTaskQuestions(questions []protocol.TaskQuestion, fallbackSubject, fallbackType, fallbackContent, fallbackAnswer string) []protocol.TaskQuestion {
	items := make([]protocol.TaskQuestion, 0, len(questions)+1)
	for _, q := range questions {
		q.Subject = strings.ToUpper(strings.TrimSpace(q.Subject))
		q.QuestionType = strings.ToUpper(strings.TrimSpace(q.QuestionType))
		q.Content = strings.TrimSpace(q.Content)
		q.Answer = strings.TrimSpace(q.Answer)
		if q.Content == "" && q.Answer == "" {
			continue
		}
		if q.Subject == "" {
			q.Subject = strings.ToUpper(strings.TrimSpace(fallbackSubject))
		}
		if q.Subject == "" {
			q.Subject = "GENERAL"
		}
		if q.QuestionType == "" {
			q.QuestionType = strings.ToUpper(strings.TrimSpace(fallbackType))
		}
		if q.QuestionType == "" {
			q.QuestionType = "NONE"
		}
		items = append(items, q)
	}
	if len(items) == 0 && (strings.TrimSpace(fallbackContent) != "" || strings.TrimSpace(fallbackAnswer) != "") {
		items = append(items, protocol.TaskQuestion{Subject: defaultTaskField(fallbackSubject, "GENERAL"), QuestionType: defaultTaskField(fallbackType, "NONE"), Content: strings.TrimSpace(fallbackContent), Answer: strings.TrimSpace(fallbackAnswer)})
	}
	return items
}

func defaultTaskField(v, fallback string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return fallback
	}
	return v
}

func encodeTaskQuestions(questions []protocol.TaskQuestion) string {
	if len(questions) == 0 {
		return "[]"
	}
	data, err := json.Marshal(questions)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeTaskQuestions(raw string, fallbackSubject, fallbackType, fallbackContent, fallbackAnswer string) []protocol.TaskQuestion {
	items := []protocol.TaskQuestion{}
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &items)
	}
	return normalizeTaskQuestions(items, fallbackSubject, fallbackType, fallbackContent, fallbackAnswer)
}
