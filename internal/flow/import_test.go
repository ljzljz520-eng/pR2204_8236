package flow

import (
	"path/filepath"
	"testing"

	"examvault/internal/domain"
	"examvault/internal/store"
)

func TestWorkflowImportReport(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "import.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := NewService(s, "import-key", 7, "importer")
	rows := []domain.ImportRow{{ID: "a", Title: "A", Permission: domain.PermissionPublic, Payload: "payload-a", Owner: "ops"}, {ID: "b", Title: "B", Permission: "invalid", Payload: "payload-b", Owner: "ops"}, {ID: "a", Title: "Duplicate", Permission: domain.PermissionStaff, Payload: "payload-c", Owner: "ops"}}
	result := service.Import(rows)
	if result.Imported != 1 || result.Rejected != 2 {
		t.Fatalf("result: %+v", result)
	}
	if len(result.IDs) != 1 || result.IDs[0] != "a" {
		t.Fatalf("ids: %+v", result.IDs)
	}
}

func TestImportRoundTrip(t *testing.T) {
	rows := []domain.ImportRow{{ID: "a", Title: "A", Permission: domain.PermissionPublic, Payload: "x", Owner: "o"}}
	decoded := ParseRows(FormatImportRows(rows))
	if len(decoded) != 1 || decoded[0].Title != "A" {
		t.Fatalf("rows: %+v", decoded)
	}
}
