package flow

import (
	"errors"
	"fmt"
	"strings"

	vaultcrypto "examvault/internal/crypto"
	"examvault/internal/domain"
	"examvault/internal/store"
)

type Service struct {
	Store          *store.Store
	Key            []byte
	Day            int
	actor          string
	lastPermission string
}

func NewService(s *store.Store, key string, day int, actor string) *Service {
	if day < 0 {
		day = 0
	}
	if strings.TrimSpace(actor) == "" {
		actor = "system"
	}
	return &Service{Store: s, Key: vaultcrypto.NormalizeKey(key), Day: day, actor: actor}
}

func (s *Service) Register(id, title, permission, payload, owner string) (domain.Record, error) {
	if s.Store == nil {
		return domain.Record{}, errors.New("service store is nil")
	}
	checksum := vaultcrypto.Digest([]byte(payload))
	record := domain.NewRecord(id, title, checksum, permission, owner, s.Day)
	if err := record.Validate(); err != nil {
		return domain.Record{}, err
	}
	ciphertext, digest, err := vaultcrypto.Encrypt(s.Key, []byte(payload))
	if err != nil {
		return domain.Record{}, err
	}
	attachment := domain.Attachment{ID: id + "-attachment", RecordID: id, Name: title + ".enc", Ciphertext: ciphertext, Digest: digest, Size: len(ciphertext), MediaType: "application/octet-stream"}
	if err := attachment.Validate(); err != nil {
		return domain.Record{}, err
	}
	workflow := domain.Workflow{ID: id + "-workflow", RecordID: id, Stage: "registration", Owner: owner, DueDay: s.Day + 7}
	if err := s.Store.PutRecord(record); err != nil {
		return domain.Record{}, err
	}
	if err := s.Store.PutAttachment(attachment); err != nil {
		return domain.Record{}, err
	}
	if err := s.Store.PutWorkflow(workflow); err != nil {
		return domain.Record{}, err
	}
	if err := s.appendAudit(id, "registered", fmt.Sprintf("title=%s", record.Title), record.Revision); err != nil {
		return domain.Record{}, err
	}
	return record, nil
}

func (s *Service) Review(id string, reviewer string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if strings.TrimSpace(reviewer) == "" {
		return record, errors.New("reviewer is empty")
	}
	if err := record.Transition(domain.StatusReviewed, s.Day); err != nil {
		return record, err
	}
	workflow, err := s.Store.GetWorkflow(id + "-workflow")
	if err == nil {
		workflow.Stage = "review"
		workflow.Owner = reviewer
		if err := s.Store.PutWorkflow(workflow); err != nil {
			return record, err
		}
	}
	if err := s.Store.PutRecord(record); err != nil {
		return record, err
	}
	if err := s.appendAudit(id, "reviewed", "reviewer="+reviewer, record.Revision); err != nil {
		return record, err
	}
	return record, nil
}

func (s *Service) Confirm(id string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if err := record.Transition(domain.StatusConfirmed, s.Day); err != nil {
		return record, err
	}
	if err := s.Store.PutRecord(record); err != nil {
		return record, err
	}
	if err := s.appendAudit(id, "confirmed", "operator="+s.actor, record.Revision); err != nil {
		return record, err
	}
	return record, nil
}

func (s *Service) appendAudit(recordID, action, detail string, revision int) error {
	sequence, err := s.Store.NextSequence()
	if err != nil {
		return err
	}
	event := domain.AuditEvent{ID: fmt.Sprintf("event-%06d", sequence), RecordID: recordID, Action: action, Actor: s.actor, Revision: revision, Detail: detail, Sequence: sequence}
	return s.Store.PutEvent(event)
}

func (s *Service) SetDay(day int) {
	if day >= 0 {
		s.Day = day
	}
}

func (s *Service) Actor() string { return s.actor }

func (s *Service) RegisterBatch(rows []domain.ImportRow) (domain.ImportResult, error) {
	if len(rows) == 0 {
		return domain.ImportResult{}, errors.New("registration batch is empty")
	}
	result := domain.ImportResult{Reasons: make([]string, 0), IDs: make([]string, 0, len(rows))}
	for _, row := range rows {
		if _, err := s.Register(row.ID, row.Title, row.Permission, row.Payload, row.Owner); err != nil {
			result.Rejected++
			result.Reasons = append(result.Reasons, row.ID+": "+err.Error())
			continue
		}
		result.Imported++
		result.IDs = append(result.IDs, row.ID)
	}
	return result, nil
}

func (s *Service) ValidateRecord(id string) error {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return err
	}
	return record.Validate()
}

func (s *Service) Attachment(id string) (domain.Attachment, error) {
	return s.Store.GetAttachment(id + "-attachment")
}

func (s *Service) Workflow(id string) (domain.Workflow, error) {
	return s.Store.GetWorkflow(id + "-workflow")
}

func (s *Service) EnsureReady(id string) (domain.Record, error) {
	record, err := s.Store.GetRecord(id)
	if err != nil {
		return record, err
	}
	if record.Status == domain.StatusArchived {
		return record, errors.New("record has been archived")
	}
	if record.Permission == domain.PermissionRestricted && record.Owner == "" {
		return record, errors.New("restricted record needs owner")
	}
	return record, nil
}
