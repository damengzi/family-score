package service

import (
	"regexp"
	"strings"
)

var dueTimePattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func normalizeDueTime(v string) string {
	v = strings.TrimSpace(v)
	if dueTimePattern.MatchString(v) {
		return v
	}
	return ""
}
