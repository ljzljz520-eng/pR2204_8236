package flow

import (
	"sort"
	"strings"

	"examvault/internal/domain"
)

type SearchOptions struct {
	Query           string
	IncludeArchived bool
	Permission      string
}

func (s *Service) Search(options SearchOptions) ([]domain.Record, error) {
	records, err := s.Store.ListRecords()
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(options.Query))
	matches := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if !options.IncludeArchived && record.Status == domain.StatusArchived {
			continue
		}
		if options.Permission != "" && record.Permission != options.Permission {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(record.Title), query) && !strings.Contains(record.Checksum, query) {
			continue
		}
		matches = append(matches, record)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Title == matches[j].Title {
			return matches[i].ID < matches[j].ID
		}
		return matches[i].Title < matches[j].Title
	})
	return matches, nil
}

func (s *Service) FindActive(id string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if record.Status == domain.StatusArchived {
		return record, domain.ErrNotFound
	}
	return record, nil
}

func (s *Service) Permissions(records []domain.Record) map[string]string {
	result := make(map[string]string, len(records))
	for _, record := range records {
		result[record.ID] = record.Permission
	}
	return result
}

func (s *Service) SearchByOwner(owner string, includeArchived bool) ([]domain.Record, error) {
	records, err := s.Store.ListRecords()
	if err != nil {
		return nil, err
	}
	result := make([]domain.Record, 0)
	for _, record := range records {
		if owner != "" && record.Owner != owner {
			continue
		}
		if !includeArchived && record.Status == domain.StatusArchived {
			continue
		}
		result = append(result, record)
	}
	domain.SortRecords(result)
	return result, nil
}

func (s *Service) SearchByStatus(status string) ([]domain.Record, error) {
	records, err := s.Store.ListRecords()
	if err != nil {
		return nil, err
	}
	result := domain.FilterByStatus(records, status)
	return result, nil
}

func FilterPermission(records []domain.Record, permission string) []domain.Record {
	result := make([]domain.Record, 0)
	for _, record := range records {
		if permission == "" || record.Permission == permission {
			result = append(result, record)
		}
	}
	domain.SortRecords(result)
	return result
}

func SortForReview(records []domain.Record) []domain.Record {
	result := domain.CloneRecords(records)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].UpdatedAt == result[j].UpdatedAt {
			return result[i].ID < result[j].ID
		}
		return result[i].UpdatedAt < result[j].UpdatedAt
	})
	return result
}
