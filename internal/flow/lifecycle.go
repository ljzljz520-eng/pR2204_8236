package flow

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"examvault/internal/domain"
)

func (s *Service) UpdatePermission(id, permission string, expectedRevision int) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if !domain.ValidPermission(permission) {
		return record, errors.New("unsupported permission")
	}
	if expectedRevision != record.Revision {
		return record, domain.ErrConflict
	}
	record.Permission = permission
	record.Revision++
	record.UpdatedAt = s.Day
	if err := record.Validate(); err != nil {
		return record, err
	}
	if err := s.Store.PutRecord(record); err != nil {
		return record, err
	}
	if err := s.appendAudit(id, "permission-updated", fmt.Sprintf("permission=%s", permission), record.Revision); err != nil {
		return record, err
	}
	return record, nil
}

func (s *Service) Publish(id string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if record.Status != domain.StatusConfirmed && record.Status != domain.StatusReviewed {
		return record, fmt.Errorf("record %s is not ready to publish", id)
	}
	if record.Status == domain.StatusReviewed {
		if err := record.Transition(domain.StatusConfirmed, s.Day); err != nil {
			return record, err
		}
	}
	if err := record.Transition(domain.StatusPublished, s.Day); err != nil {
		return record, err
	}
	if err := s.Store.PutRecord(record); err != nil {
		return record, err
	}
	if err := s.appendAudit(id, "published", "operator="+s.actor, record.Revision); err != nil {
		return record, err
	}
	return record, nil
}

func (s *Service) Archive(id string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if record.Status != domain.StatusPublished {
		return record, errors.New("only published records can be archived")
	}
	if err := record.Transition(domain.StatusArchived, s.Day); err != nil {
		return record, err
	}
	if err := s.Store.PutRecord(record); err != nil {
		return record, err
	}
	workflow, err := s.Store.GetWorkflow(id + "-workflow")
	if err == nil {
		workflow.Stage = "archive"
		workflow.Completed = true
		if err := s.Store.PutWorkflow(workflow); err != nil {
			return record, err
		}
	}
	if err := s.appendAudit(id, "archived", "operator="+s.actor, record.Revision); err != nil {
		return record, err
	}
	return record, nil
}

func (s *Service) Refresh(id string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if strings.TrimSpace(s.lastPermission) == "" {
		s.lastPermission = record.Permission
	} else {
		record.Permission = s.lastPermission
		record.Revision++
		record.UpdatedAt = s.Day
		if err := s.Store.PutRecord(record); err != nil {
			return record, err
		}
	}
	return record, nil
}

func (s *Service) ResetRefresh() { s.lastPermission = "" }

func (s *Service) BulkReview(ids []string, reviewer string) ([]domain.Record, []string) {
	reviewed := make([]domain.Record, 0, len(ids))
	errors := make([]string, 0)
	for _, id := range ids {
		record, err := s.Review(id, reviewer)
		if err != nil {
			errors = append(errors, id+": "+err.Error())
			continue
		}
		reviewed = append(reviewed, record)
	}
	return reviewed, errors
}

func (s *Service) ConfirmAll(ids []string) ([]domain.Record, []string) {
	confirmed := make([]domain.Record, 0, len(ids))
	errors := make([]string, 0)
	for _, id := range ids {
		record, err := s.Confirm(id)
		if err != nil {
			errors = append(errors, id+": "+err.Error())
			continue
		}
		confirmed = append(confirmed, record)
	}
	return confirmed, errors
}

func (s *Service) PublishAll(ids []string) ([]domain.Record, []string) {
	published := make([]domain.Record, 0, len(ids))
	errors := make([]string, 0)
	for _, id := range ids {
		record, err := s.Publish(id)
		if err != nil {
			errors = append(errors, id+": "+err.Error())
			continue
		}
		published = append(published, record)
	}
	return published, errors
}

func (s *Service) ArchiveAll(ids []string) ([]domain.Record, []string) {
	archived := make([]domain.Record, 0, len(ids))
	errors := make([]string, 0)
	for _, id := range ids {
		record, err := s.Archive(id)
		if err != nil {
			errors = append(errors, id+": "+err.Error())
			continue
		}
		archived = append(archived, record)
	}
	return archived, errors
}

func (s *Service) UpdateMany(changes map[string]string) (map[string]domain.Record, []string) {
	updated := make(map[string]domain.Record, len(changes))
	errors := make([]string, 0)
	ids := make([]string, 0, len(changes))
	for id := range changes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		record, err := s.Store.GetRecord(id)
		if err != nil {
			errors = append(errors, id+": "+err.Error())
			continue
		}
		changed, err := s.UpdatePermission(id, changes[id], record.Revision)
		if err != nil {
			errors = append(errors, id+": "+err.Error())
			continue
		}
		updated[id] = changed
	}
	return updated, errors
}
