package report

import (
	"testing"

	"examvault/internal/domain"
)

func TestReportSummaries(t *testing.T) {
	records := []domain.Record{{ID: "b", Title: "B", Permission: domain.PermissionStaff, Status: domain.StatusReviewed}, {ID: "a", Title: "A", Permission: domain.PermissionPublic, Status: domain.StatusDraft}}
	summary := SearchSummary("", records)
	if summary.Count != 2 || summary.IDs[0] != "a" {
		t.Fatalf("summary: %+v", summary)
	}
	permissions := PermissionSummary(records)
	if permissions[domain.PermissionPublic] != 1 {
		t.Fatal(permissions)
	}
}

func TestReportEncoding(t *testing.T) {
	if EncodeSearch(SearchSummary("x", nil)) == "" || EncodeImport(ImportReport{}) == "" {
		t.Fatal("encoding failed")
	}
}
