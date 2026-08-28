package crypto

import "testing"

func TestEncryptDeterministic(t *testing.T) {
	key := []byte("fixed-key")
	one, digest, err := Encrypt(key, []byte("exam paper"))
	if err != nil {
		t.Fatal(err)
	}
	two, digest2, err := Encrypt(key, []byte("exam paper"))
	if err != nil {
		t.Fatal(err)
	}
	if string(one) != string(two) || digest != digest2 {
		t.Fatal("encryption is not deterministic")
	}
	plain, err := Decrypt(key, one, digest)
	if err != nil || string(plain) != "exam paper" {
		t.Fatalf("decrypt: %s %v", plain, err)
	}
}

func TestDigestVerification(t *testing.T) {
	if !Verify([]byte("x"), Digest([]byte("x"))) {
		t.Fatal("digest should verify")
	}
}
