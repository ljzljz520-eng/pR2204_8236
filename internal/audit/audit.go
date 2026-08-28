package audit

import (
	"fmt"
	"sort"
	"strings"

	"examvault/internal/domain"
	"examvault/internal/store"
)

type Reader struct{ Store *store.Store }

func NewReader(s *store.Store) Reader { return Reader{Store: s} }

func (r Reader) List(recordID string) ([]domain.AuditEvent, error) {
	events, err := r.Store.ListEvents(recordID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Sequence < events[j].Sequence })
	return events, nil
}

func (r Reader) Actions(recordID string) ([]string, error) {
	events, err := r.List(recordID)
	if err != nil {
		return nil, err
	}
	actions := make([]string, 0, len(events))
	for _, event := range events {
		actions = append(actions, event.Action)
	}
	return actions, nil
}

func Format(event domain.AuditEvent) string {
	return fmt.Sprintf("%04d %s %s r%d %s", event.Sequence, event.Actor, event.Action, event.Revision, event.Detail)
}

func FormatAll(events []domain.AuditEvent) string {
	lines := make([]string, 0, len(events))
	for _, event := range events {
		lines = append(lines, Format(event))
	}
	return strings.Join(lines, "\n")
}

type Summary struct {
	Total      int
	ByAction   map[string]int
	ByActor    map[string]int
	LastAction string
}

func Summarize(events []domain.AuditEvent) Summary {
	summary := Summary{ByAction: make(map[string]int), ByActor: make(map[string]int)}
	ordered := append([]domain.AuditEvent(nil), events...)
	domain.SortEvents(ordered)
	for _, event := range ordered {
		summary.Total++
		summary.ByAction[event.Action]++
		summary.ByActor[event.Actor]++
		summary.LastAction = event.Action
	}
	return summary
}

func (r Reader) ByActor(recordID, actor string) ([]domain.AuditEvent, error) {
	events, err := r.List(recordID)
	if err != nil {
		return nil, err
	}
	return FilterEvents(events, Filter{Actor: actor}), nil
}

func (r Reader) ByAction(recordID, action string) ([]domain.AuditEvent, error) {
	events, err := r.List(recordID)
	if err != nil {
		return nil, err
	}
	return FilterEvents(events, Filter{Action: action}), nil
}

func (r Reader) Since(recordID string, sequence int) ([]domain.AuditEvent, error) {
	events, err := r.List(recordID)
	if err != nil {
		return nil, err
	}
	return FilterEvents(events, Filter{FromSequence: sequence}), nil
}

func (r Reader) Format(recordID string) (string, error) {
	events, err := r.List(recordID)
	if err != nil {
		return "", err
	}
	return FormatAll(events), nil
}

func ActionsByRevision(events []domain.AuditEvent) map[int][]string {
	result := make(map[int][]string)
	for _, event := range events {
		result[event.Revision] = append(result[event.Revision], event.Action)
	}
	return result
}

func SequenceRange(events []domain.AuditEvent) (int, int) {
	if len(events) == 0 {
		return 0, 0
	}
	domain.SortEvents(events)
	return events[0].Sequence, events[len(events)-1].Sequence
}
