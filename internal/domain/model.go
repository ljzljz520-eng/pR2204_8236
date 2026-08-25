package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	PermissionPublic     = "public"
	PermissionStaff      = "staff"
	PermissionRestricted = "restricted"
)

const (
	StatusDraft     = "draft"
	StatusReviewed  = "reviewed"
	StatusConfirmed = "confirmed"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

type Record struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Checksum   string `json:"checksum"`
	Permission string `json:"permission"`
	Status     string `json:"status"`
	Revision   int    `json:"revision"`
	CreatedAt  int    `json:"created_at"`
	UpdatedAt  int    `json:"updated_at"`
	Owner      string `json:"owner"`
	Notes      string `json:"notes"`
}

type AuditEvent struct {
	ID       string `json:"id"`
	RecordID string `json:"record_id"`
	Action   string `json:"action"`
	Actor    string `json:"actor"`
	Revision int    `json:"revision"`
	Detail   string `json:"detail"`
	Sequence int    `json:"sequence"`
}

type Workflow struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Stage     string `json:"stage"`
	Owner     string `json:"owner"`
	DueDay    int    `json:"due_day"`
	Completed bool   `json:"completed"`
}

type Attachment struct {
	ID         string `json:"id"`
	RecordID   string `json:"record_id"`
	Name       string `json:"name"`
	Ciphertext []byte `json:"ciphertext"`
	Digest     string `json:"digest"`
	Size       int    `json:"size"`
	MediaType  string `json:"media_type"`
}

type ImportRow struct {
	ID         string
	Title      string
	Permission string
	Payload    string
	Owner      string
}

type ImportResult struct {
	Imported int
	Rejected int
	Reasons  []string
	IDs      []string
}

func NewRecord(id, title, checksum, permission, owner string, day int) Record {
	return Record{ID: id, Title: strings.TrimSpace(title), Checksum: strings.ToLower(strings.TrimSpace(checksum)), Permission: permission, Status: StatusDraft, Revision: 1, CreatedAt: day, UpdatedAt: day, Owner: owner}
}

func (r Record) Clone() Record { return r }

func (r Record) IsActive() bool { return r.Status != StatusArchived }

func (r Record) Marshal() ([]byte, error) { return json.Marshal(r) }

func DecodeRecord(data []byte) (Record, error) {
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return r, err
	}
	return r, nil
}

func (e AuditEvent) Marshal() ([]byte, error) { return json.Marshal(e) }

func DecodeAuditEvent(data []byte) (AuditEvent, error) {
	var e AuditEvent
	if err := json.Unmarshal(data, &e); err != nil {
		return e, err
	}
	return e, nil
}

func (w Workflow) Marshal() ([]byte, error) { return json.Marshal(w) }

func DecodeWorkflow(data []byte) (Workflow, error) {
	var w Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		return w, err
	}
	return w, nil
}

func (a Attachment) Marshal() ([]byte, error) { return json.Marshal(a) }

func DecodeAttachment(data []byte) (Attachment, error) {
	var a Attachment
	if err := json.Unmarshal(data, &a); err != nil {
		return a, err
	}
	return a, nil
}

func SortRecords(records []Record) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].ID == records[j].ID {
			return records[i].Revision < records[j].Revision
		}
		return records[i].ID < records[j].ID
	})
}

func SortEvents(events []AuditEvent) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].Sequence == events[j].Sequence {
			return events[i].ID < events[j].ID
		}
		return events[i].Sequence < events[j].Sequence
	})
}

var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("revision conflict")
	ErrInvalidTransition = errors.New("invalid status transition")
)

func WrapInvalid(field string) error { return fmt.Errorf("invalid %s", field) }

func AllPermissions() []string {
	return []string{PermissionPublic, PermissionStaff, PermissionRestricted}
}

func PermissionRank(permission string) int {
	switch permission {
	case PermissionPublic:
		return 1
	case PermissionStaff:
		return 2
	case PermissionRestricted:
		return 3
	default:
		return 0
	}
}

func PermissionLabel(permission string) string {
	switch permission {
	case PermissionPublic:
		return "Public"
	case PermissionStaff:
		return "Staff only"
	case PermissionRestricted:
		return "Restricted"
	default:
		return "Unknown"
	}
}

func StatusOrder(status string) int {
	switch status {
	case StatusDraft:
		return 1
	case StatusReviewed:
		return 2
	case StatusConfirmed:
		return 3
	case StatusPublished:
		return 4
	case StatusArchived:
		return 5
	default:
		return 0
	}
}

func StatusLabel(status string) string {
	switch status {
	case StatusDraft:
		return "Draft"
	case StatusReviewed:
		return "Reviewed"
	case StatusConfirmed:
		return "Confirmed"
	case StatusPublished:
		return "Published"
	case StatusArchived:
		return "Archived"
	default:
		return "Unknown"
	}
}

func TransitionPath(from, to string) []string {
	if !ValidStatus(from) || !ValidStatus(to) {
		return nil
	}
	orderFrom, orderTo := StatusOrder(from), StatusOrder(to)
	if orderFrom > orderTo {
		return nil
	}
	all := []string{StatusDraft, StatusReviewed, StatusConfirmed, StatusPublished, StatusArchived}
	path := make([]string, 0, orderTo-orderFrom)
	for index := orderFrom; index < orderTo; index++ {
		path = append(path, all[index])
	}
	return path
}

func ValidateCollection(records []Record) error {
	seen := make(map[string]bool, len(records))
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if seen[record.ID] {
			return fmt.Errorf("duplicate record %s", record.ID)
		}
		seen[record.ID] = true
	}
	return nil
}

func CloneRecords(records []Record) []Record {
	clones := make([]Record, len(records))
	copy(clones, records)
	return clones
}

func GroupRecordsByStatus(records []Record) map[string][]Record {
	groups := make(map[string][]Record)
	for _, record := range records {
		groups[record.Status] = append(groups[record.Status], record)
	}
	for status, grouped := range groups {
		SortRecords(grouped)
		groups[status] = grouped
	}
	return groups
}

func GroupRecordsByPermission(records []Record) map[string][]Record {
	groups := make(map[string][]Record)
	for _, record := range records {
		groups[record.Permission] = append(groups[record.Permission], record)
	}
	for permission, grouped := range groups {
		SortRecords(grouped)
		groups[permission] = grouped
	}
	return groups
}

func FilterByOwner(records []Record, owner string) []Record {
	result := make([]Record, 0)
	for _, record := range records {
		if owner == "" || record.Owner == owner {
			result = append(result, record)
		}
	}
	SortRecords(result)
	return result
}

func FilterByStatus(records []Record, status string) []Record {
	result := make([]Record, 0)
	for _, record := range records {
		if status == "" || record.Status == status {
			result = append(result, record)
		}
	}
	SortRecords(result)
	return result
}

func DueDayFor(status string, created int) int {
	delta := 3
	switch status {
	case StatusDraft:
		delta = 2
	case StatusReviewed:
		delta = 4
	case StatusConfirmed:
		delta = 5
	case StatusPublished:
		delta = 30
	case StatusArchived:
		delta = 0
	}
	return created + delta
}

func NormalizeTitle(title string) string {
	parts := strings.Fields(strings.TrimSpace(title))
	return strings.Join(parts, " ")
}

func RecordKey(record Record) string { return record.ID + "@" + record.Checksum }

func EventSummary(event AuditEvent) string {
	return fmt.Sprintf("%s:%s:%d:%s", event.RecordID, event.Action, event.Revision, event.Actor)
}

func AttachmentSummary(attachment Attachment) string {
	return fmt.Sprintf("%s/%s:%d:%s", attachment.RecordID, attachment.Name, attachment.Size, attachment.Digest)
}
