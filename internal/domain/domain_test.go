package domain

import "testing"

func TestRecordValidationAndTransition(t *testing.T) {
	r := NewRecord("r1", "Paper", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PermissionStaff, "alice", 1)
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := r.Transition(StatusReviewed, 2); err != nil {
		t.Fatal(err)
	}
	if r.Status != StatusReviewed || r.Revision != 2 {
		t.Fatalf("unexpected record: %+v", r)
	}
	if err := r.Transition(StatusArchived, 3); err == nil {
		t.Fatal("expected transition error")
	}
}

func TestAttachmentValidation(t *testing.T) {
	a := Attachment{ID: "a", RecordID: "r", Name: "x", Ciphertext: []byte{1, 2}, Digest: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Size: 2}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	a.Size = 4
	if err := a.Validate(); err == nil {
		t.Fatal("expected size error")
	}
}
