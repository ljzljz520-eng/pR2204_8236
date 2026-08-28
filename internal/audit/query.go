package audit

import (
	"sort"
	"strings"

	"examvault/internal/domain"
)

type Filter struct {
	Actor        string
	Action       string
	FromSequence int
	ToSequence   int
}

func FilterEvents(events []domain.AuditEvent, filter Filter) []domain.AuditEvent {
	result := make([]domain.AuditEvent, 0, len(events))
	for _, event := range events {
		if filter.Actor != "" && event.Actor != filter.Actor {
			continue
		}
		if filter.Action != "" && event.Action != filter.Action {
			continue
		}
		if filter.FromSequence > 0 && event.Sequence < filter.FromSequence {
			continue
		}
		if filter.ToSequence > 0 && event.Sequence > filter.ToSequence {
			continue
		}
		result = append(result, event)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Sequence < result[j].Sequence })
	return result
}

func ContainsAction(events []domain.AuditEvent, action string) bool {
	target := strings.TrimSpace(action)
	for _, event := range events {
		if event.Action == target {
			return true
		}
	}
	return false
}

func Latest(events []domain.AuditEvent) (domain.AuditEvent, bool) {
	if len(events) == 0 {
		return domain.AuditEvent{}, false
	}
	latest := events[0]
	for _, event := range events[1:] {
		if event.Sequence > latest.Sequence {
			latest = event
		}
	}
	return latest, true
}

func CountByAction(events []domain.AuditEvent) map[string]int {
	counts := make(map[string]int)
	for _, event := range events {
		counts[event.Action]++
	}
	return counts
}

func CountByRecord(events []domain.AuditEvent) map[string]int {
	counts := make(map[string]int)
	for _, event := range events {
		counts[event.RecordID]++
	}
	return counts
}

func Reverse(events []domain.AuditEvent) []domain.AuditEvent {
	result := append([]domain.AuditEvent(nil), events...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func ActionsText(events []domain.AuditEvent) string {
	actions := make([]string, 0, len(events))
	for _, event := range events {
		actions = append(actions, event.Action)
	}
	return strings.Join(actions, ",")
}
