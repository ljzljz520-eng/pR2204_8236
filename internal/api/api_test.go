package api

import (
	"path/filepath"
	"testing"

	"examvault/internal/flow"
	"examvault/internal/store"
)

func TestCLIParse(t *testing.T) {
	command, err := ParseArgs([]string{"search", "paper"})
	if err != nil || command.Name != "search" || len(command.Args) != 1 {
		t.Fatalf("command: %+v %v", command, err)
	}
	if _, err := ParseArgs([]string{"unknown"}); err == nil {
		t.Fatal("expected unknown command")
	}
}

func TestCLIRun(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "cli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runner := Runner{Service: flow.NewService(s, "cli-key", 1, "cli")}
	output, err := runner.Run([]string{"register", "r", "Paper", "public", "payload", "owner"})
	if err != nil || output == "" {
		t.Fatalf("run: %s %v", output, err)
	}
}
