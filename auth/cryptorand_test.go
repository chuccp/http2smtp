package auth

import (
	"testing"

	"github.com/chuccp/http2smtp/entity"
)

// The session cipher keys come from crypto/rand; this pins the round trip that every
// login depends on, and that the key material is not degenerate.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	in := &entity.LoginUser{Id: 7, Username: "alice", Salt: "salt", IsAdmin: true}
	sealed, err := Encrypt(in)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	var out entity.LoginUser
	if err := Decrypt(sealed, &out); err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if out.Id != in.Id || out.Username != in.Username || out.Salt != in.Salt || out.IsAdmin != in.IsAdmin {
		t.Errorf("round trip mismatch: %+v", out)
	}
	if len(aesKey) != 32 {
		t.Errorf("aesKey length = %d, want 32", len(aesKey))
	}
	if len(aesIv) != 16 {
		t.Errorf("aesIv length = %d, want 16", len(aesIv))
	}
}

// Two draws must differ: a fixed key would make every process's cookies interchangeable.
func TestCryptoRandomAlnumVaries(t *testing.T) {
	a := cryptoRandomAlnum(32)
	b := cryptoRandomAlnum(32)
	if a == b {
		t.Error("two independent draws produced the same string")
	}
	for _, r := range a {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			t.Fatalf("character %q outside the alphanumeric charset", r)
		}
	}
}
