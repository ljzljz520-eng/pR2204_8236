package flow007

import (
	"path/filepath"
	"testing"

	"examvault/internal/domain"
	"examvault/internal/flow"
	"examvault/internal/store"
)

func Test2204BusinessRegression(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "regression.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := flow.NewService(database, "regression-key", 9, "reviewer")
	if _, err = service.Register("paper-a", "Morning Exam", domain.PermissionStaff, "A", "ops"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Register("paper-b", "Afternoon Exam", domain.PermissionRestricted, "B", "ops"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Refresh("paper-a"); err != nil {
		t.Fatal(err)
	}
	refreshed, err := service.Refresh("paper-b")
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.Permission != domain.PermissionRestricted {
		t.Fatalf("paper-b permission = %s", refreshed.Permission)
	}
}
