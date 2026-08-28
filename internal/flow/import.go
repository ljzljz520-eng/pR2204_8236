package flow

import (
	"fmt"
	"strings"

	vaultcrypto "examvault/internal/crypto"
	"examvault/internal/domain"
	"examvault/internal/store"
)

func (s *Service) Import(rows []domain.ImportRow) domain.ImportResult {
	result := domain.ImportResult{Reasons: make([]string, 0), IDs: make([]string, 0)}
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		if err := domain.EnsureImportRow(row); err != nil {
			result.Rejected++
			result.Reasons = append(result.Reasons, row.ID+": "+err.Error())
			continue
		}
		if seen[row.ID] {
			result.Rejected++
			result.Reasons = append(result.Reasons, row.ID+": duplicate")
			continue
		}
		seen[row.ID] = true
		if _, err := s.Store.GetRecord(row.ID); err == nil {
			result.Rejected++
			result.Reasons = append(result.Reasons, row.ID+": exists")
			continue
		}
		record, err := s.Register(row.ID, row.Title, row.Permission, row.Payload, row.Owner)
		if err != nil {
			result.Rejected++
			result.Reasons = append(result.Reasons, row.ID+": "+err.Error())
			continue
		}
		result.Imported++
		result.IDs = append(result.IDs, record.ID)
	}
	return result
}

func ParseRows(raw string) []domain.ImportRow { return store.DecodeImportRows(raw) }

func FormatImportRows(rows []domain.ImportRow) string { return store.EncodeImportRows(rows) }

func (s *Service) ValidateImport(rows []domain.ImportRow) []string {
	errors := make([]string, 0)
	for _, row := range rows {
		if err := domain.EnsureImportRow(row); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", row.ID, err))
		}
	}
	return errors
}

func (s *Service) ImportDigest(rows []domain.ImportRow) string {
	var builder strings.Builder
	for _, row := range rows {
		builder.WriteString(row.ID)
		builder.WriteByte(':')
		builder.WriteString(vaultcrypto.Digest([]byte(row.Payload)))
		builder.WriteByte(';')
	}
	return vaultcrypto.Digest([]byte(builder.String()))
}

func NormalizeRows(rows []domain.ImportRow) []domain.ImportRow {
	result := make([]domain.ImportRow, 0, len(rows))
	for _, row := range rows {
		row.ID = strings.TrimSpace(row.ID)
		row.Title = domain.NormalizeTitle(row.Title)
		row.Permission = domain.NormalizePermission(row.Permission)
		row.Owner = strings.TrimSpace(row.Owner)
		result = append(result, row)
	}
	return result
}

func RejectReasons(rows []domain.ImportRow) map[string]string {
	result := make(map[string]string)
	for _, row := range rows {
		if err := domain.EnsureImportRow(row); err != nil {
			result[row.ID] = err.Error()
		}
	}
	return result
}

func SplitRows(rows []domain.ImportRow) (valid []domain.ImportRow, invalid []domain.ImportRow) {
	for _, row := range rows {
		if domain.EnsureImportRow(row) == nil {
			valid = append(valid, row)
		} else {
			invalid = append(invalid, row)
		}
	}
	return valid, invalid
}
