package store

import (
	"path/filepath"
	"testing"

	"examvault/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.db")
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := domain.NewRecord("r1", "Algebra", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", domain.PermissionRestricted, "exam", 2)
	if err := first.PutRecord(record); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetRecord("r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Algebra" || got.Permission != domain.PermissionRestricted {
		t.Fatalf("unexpected reopened record: %+v", got)
	}
}

func TestStoreCollections(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "records.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	w := domain.Workflow{ID: "w", RecordID: "r", Stage: "registration", Owner: "a", DueDay: 3}
	if err := s.PutWorkflow(w); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetWorkflow("w"); err != nil || got.Stage != w.Stage {
		t.Fatalf("workflow mismatch: %+v %v", got, err)
	}
	seq, err := s.NextSequence()
	if err != nil || seq != 1 {
		t.Fatalf("sequence: %d %v", seq, err)
	}
}
