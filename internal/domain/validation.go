package domain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

func (r Record) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return WrapInvalid("id")
	}
	if strings.TrimSpace(r.Title) == "" {
		return WrapInvalid("title")
	}
	if len(r.Checksum) != 64 {
		return WrapInvalid("checksum")
	}
	if _, err := hex.DecodeString(r.Checksum); err != nil {
		return WrapInvalid("checksum")
	}
	if !ValidPermission(r.Permission) {
		return WrapInvalid("permission")
	}
	if !ValidStatus(r.Status) {
		return WrapInvalid("status")
	}
	if r.Revision < 1 {
		return WrapInvalid("revision")
	}
	if r.CreatedAt < 0 || r.UpdatedAt < r.CreatedAt {
		return WrapInvalid("dates")
	}
	if strings.TrimSpace(r.Owner) == "" {
		return WrapInvalid("owner")
	}
	return nil
}

func ValidPermission(value string) bool {
	switch value {
	case PermissionPublic, PermissionStaff, PermissionRestricted:
		return true
	default:
		return false
	}
}

func ValidStatus(value string) bool {
	switch value {
	case StatusDraft, StatusReviewed, StatusConfirmed, StatusPublished, StatusArchived:
		return true
	default:
		return false
	}
}

func (a Attachment) Validate() error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.RecordID) == "" {
		return WrapInvalid("attachment identity")
	}
	if strings.TrimSpace(a.Name) == "" {
		return WrapInvalid("attachment name")
	}
	if a.Size != len(a.Ciphertext) {
		return WrapInvalid("attachment size")
	}
	if len(a.Digest) != 64 {
		return WrapInvalid("attachment digest")
	}
	if _, err := hex.DecodeString(a.Digest); err != nil {
		return WrapInvalid("attachment digest")
	}
	return nil
}

func (r Record) CanTransition(next string) bool {
	switch r.Status {
	case StatusDraft:
		return next == StatusReviewed
	case StatusReviewed:
		return next == StatusConfirmed || next == StatusDraft
	case StatusConfirmed:
		return next == StatusPublished
	case StatusPublished:
		return next == StatusArchived
	case StatusArchived:
		return false
	default:
		return false
	}
}

func (r *Record) Transition(next string, day int) error {
	if !r.CanTransition(next) {
		return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, r.Status, next)
	}
	r.Status = next
	r.UpdatedAt = day
	r.Revision++
	return nil
}

func EnsureImportRow(row ImportRow) error {
	if strings.TrimSpace(row.ID) == "" {
		return errors.New("missing id")
	}
	if strings.TrimSpace(row.Title) == "" {
		return errors.New("missing title")
	}
	if !ValidPermission(row.Permission) {
		return errors.New("invalid permission")
	}
	if strings.TrimSpace(row.Payload) == "" {
		return errors.New("missing payload")
	}
	if strings.TrimSpace(row.Owner) == "" {
		return errors.New("missing owner")
	}
	return nil
}

func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is empty")
	}
	if len([]rune(title)) > 160 {
		return errors.New("title is too long")
	}
	return nil
}

func ValidateOwner(owner string) error {
	if strings.TrimSpace(owner) == "" {
		return errors.New("owner is empty")
	}
	if strings.ContainsAny(owner, "\n\r|") {
		return errors.New("owner contains separators")
	}
	return nil
}

func ValidateChecksum(checksum string) error {
	clean := strings.ToLower(strings.TrimSpace(checksum))
	if len(clean) != 64 {
		return errors.New("checksum length is invalid")
	}
	if _, err := hex.DecodeString(clean); err != nil {
		return errors.New("checksum encoding is invalid")
	}
	return nil
}

func NormalizePermission(permission string) string {
	clean := strings.ToLower(strings.TrimSpace(permission))
	switch clean {
	case "open":
		return PermissionPublic
	case "internal":
		return PermissionStaff
	case "private":
		return PermissionRestricted
	default:
		return clean
	}
}

func NormalizeStatus(status string) string { return strings.ToLower(strings.TrimSpace(status)) }

func ValidateRecordForWrite(record Record) error {
	if err := ValidateTitle(record.Title); err != nil {
		return err
	}
	if err := ValidateOwner(record.Owner); err != nil {
		return err
	}
	if err := ValidateChecksum(record.Checksum); err != nil {
		return err
	}
	return record.Validate()
}

func IsTerminal(status string) bool { return status == StatusArchived }

func IsReviewable(status string) bool {
	return status == StatusDraft || status == StatusReviewed
}

func IsPublishable(status string) bool {
	return status == StatusReviewed || status == StatusConfirmed
}

func IsEditable(status string) bool { return status != StatusArchived && status != StatusPublished }
