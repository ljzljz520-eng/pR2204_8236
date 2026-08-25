package flow

import (
	"path/filepath"
	"testing"

	"examvault/internal/domain"
	"examvault/internal/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "flow.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return NewService(s, "key", 4, "tester")
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	service := testService(t)
	record, err := service.Register("r1", "Physics", domain.PermissionStaff, "encrypted payload", "owner")
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != domain.StatusDraft {
		t.Fatal(record.Status)
	}
	if _, err = service.Review("r1", "reviewer"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Confirm("r1"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Publish("r1"); err != nil {
		t.Fatal(err)
	}
	archived, err := service.Archive("r1")
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != domain.StatusArchived {
		t.Fatal(archived.Status)
	}
	events, err := service.Timeline("r1")
	if err != nil || len(events) != 5 {
		t.Fatalf("events: %d %v", len(events), err)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	service := testService(t)
	first, err := service.Register("r1", "Chemistry", domain.PermissionPublic, "one", "owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Register("r2", "History", domain.PermissionRestricted, "two", "owner"); err != nil {
		t.Fatal(err)
	}
	found, err := service.Search(SearchOptions{Query: "chem"})
	if err != nil || len(found) != 1 {
		t.Fatalf("search: %d %v", len(found), err)
	}
	updated, err := service.UpdatePermission(first.ID, domain.PermissionStaff, first.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Permission != domain.PermissionStaff || updated.Revision != 2 {
		t.Fatalf("updated: %+v", updated)
	}
	if _, err = service.Publish(first.ID); err == nil {
		t.Fatal("draft should not publish")
	}
	if _, err = service.Review(first.ID, "reviewer"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Publish(first.ID); err != nil {
		t.Fatal(err)
	}
}
