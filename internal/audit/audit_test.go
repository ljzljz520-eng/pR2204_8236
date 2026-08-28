package audit

import (
	"path/filepath"
	"testing"

	"examvault/internal/domain"
	"examvault/internal/flow"
	"examvault/internal/store"
)

func TestAuditHistory(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := flow.NewService(s, "audit-key", 1, "auditor")
	if _, err = service.Register("r", "Report", domain.PermissionPublic, "x", "owner"); err != nil {
		t.Fatal(err)
	}
	reader := NewReader(s)
	events, err := reader.List("r")
	if err != nil || len(events) != 1 {
		t.Fatalf("events: %+v %v", events, err)
	}
	if !ContainsAction(events, "registered") {
		t.Fatal("missing action")
	}
	if _, ok := Latest(events); !ok {
		t.Fatal("latest missing")
	}
}
