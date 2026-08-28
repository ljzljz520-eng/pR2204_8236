package flow

import (
	"fmt"
	"strings"

	"examvault/internal/domain"
)

func (s *Service) Timeline(recordID string) ([]domain.AuditEvent, error) {
	return s.Store.ListEvents(recordID)
}

func (s *Service) StatusCounts(includeArchived bool) (map[string]int, error) {
	records, err := s.Store.ListRecords()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, record := range records {
		if record.Status == domain.StatusArchived && !includeArchived {
			continue
		}
		counts[record.Status]++
	}
	return counts, nil
}

func Explain(record domain.Record) string {
	return fmt.Sprintf("%s [%s] permission=%s revision=%d owner=%s", record.ID, record.Status, record.Permission, record.Revision, record.Owner)
}

func JoinReasons(result domain.ImportResult) string { return strings.Join(result.Reasons, "; ") }

func (s *Service) AuditDigest(recordID string) (string, error) {
	events, err := s.Timeline(recordID)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(events))
	for _, event := range events {
		parts = append(parts, domain.EventSummary(event))
	}
	return strings.Join(parts, "|"), nil
}

func (s *Service) ReviewQueue() ([]domain.Record, error) {
	records, err := s.Search(SearchOptions{IncludeArchived: false})
	if err != nil {
		return nil, err
	}
	queue := make([]domain.Record, 0)
	for _, record := range records {
		if record.Status == domain.StatusDraft || record.Status == domain.StatusReviewed {
			queue = append(queue, record)
		}
	}
	return queue, nil
}

func (s *Service) OwnershipReport() (map[string]int, error) {
	records, err := s.Store.ListRecords()
	if err != nil {
		return nil, err
	}
	owners := make(map[string]int)
	for _, record := range records {
		owners[record.Owner]++
	}
	return owners, nil
}
