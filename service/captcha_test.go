package service

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newTestCaptchaService(now func() time.Time) *CaptchaService {
	return &CaptchaService{challenges: make(map[string]*captchaChallenge), now: now, key: []byte("test-key")}
}

// captchaPoints reads the expected clicks before Verify consumes the challenge.
func captchaPoints(t *testing.T, s *CaptchaService, id string) (xs, ys []int) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge := s.challenges[id]
	if challenge == nil {
		t.Fatalf("challenge %q is not stored", id)
	}
	for _, target := range challenge.targets {
		xs = append(xs, target.x)
		ys = append(ys, target.y)
	}
	return xs, ys
}

func TestCaptchaServiceChallenge(t *testing.T) {
	s := newTestCaptchaService(time.Now)
	data, err := s.New("alice", "secret")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if data.ID == "" {
		t.Error("challenge id is empty")
	}
	if len(data.Shapes) != 3 {
		t.Errorf("shapes = %v, want 3 in click order", data.Shapes)
	}
	if data.Width != captchaWidth || data.Height != captchaHeight {
		t.Errorf("size = %dx%d, want %dx%d", data.Width, data.Height, captchaWidth, captchaHeight)
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(data.Image, prefix) {
		t.Fatalf("image is not a data URI: %.30s", data.Image)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(data.Image, prefix))
	if err != nil {
		t.Fatalf("decoding image: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("rendered image is not PNG: %v", err)
	}
	if img.Bounds().Dx() != captchaWidth || img.Bounds().Dy() != captchaHeight {
		t.Errorf("image bounds = %v", img.Bounds())
	}

	xs, ys := captchaPoints(t, s, data.ID)
	// A click within the tolerance of each target, in prompt order, passes.
	xsn := append([]int(nil), xs...)
	ysn := append([]int(nil), ys...)
	for i := range xsn {
		xsn[i] += 5
		ysn[i] -= 5
	}
	if !s.Verify(data.ID, "alice", "secret", xsn, ysn) {
		t.Error("the expected clicks were rejected")
	}
	// The challenge is one-shot: the same answer cannot be replayed.
	if s.Verify(data.ID, "alice", "secret", xsn, ysn) {
		t.Error("a consumed challenge accepted a replay")
	}
}

func TestCaptchaServiceRejectsWrongClicks(t *testing.T) {
	s := newTestCaptchaService(time.Now)

	t.Run("wrong order", func(t *testing.T) {
		data, err := s.New("alice", "secret")
		if err != nil {
			t.Fatal(err)
		}
		xs, ys := captchaPoints(t, s, data.ID)
		xs[0], xs[1] = xs[1], xs[0]
		ys[0], ys[1] = ys[1], ys[0]
		if s.Verify(data.ID, "alice", "secret", xs, ys) {
			t.Error("a reversed click order was accepted")
		}
	})

	t.Run("off target", func(t *testing.T) {
		data, err := s.New("alice", "secret")
		if err != nil {
			t.Fatal(err)
		}
		xs, ys := captchaPoints(t, s, data.ID)
		xs[0] += captchaTolerance + 1
		if s.Verify(data.ID, "alice", "secret", xs, ys) {
			t.Error("a click outside the tolerance was accepted")
		}
	})

	t.Run("wrong number of clicks", func(t *testing.T) {
		data, err := s.New("alice", "secret")
		if err != nil {
			t.Fatal(err)
		}
		xs, ys := captchaPoints(t, s, data.ID)
		if s.Verify(data.ID, "alice", "secret", xs[:2], ys[:2]) {
			t.Error("too few clicks were accepted")
		}
	})

	t.Run("unknown id", func(t *testing.T) {
		if s.Verify("no-such-id", "alice", "secret", []int{1, 2, 3}, []int{1, 2, 3}) {
			t.Error("an unknown challenge id was accepted")
		}
		if s.Verify("", "alice", "secret", nil, nil) {
			t.Error("an empty challenge id was accepted")
		}
	})
}

func TestCaptchaServiceExpires(t *testing.T) {
	now := time.Now()
	s := newTestCaptchaService(func() time.Time { return now })
	data, err := s.New("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	xs, ys := captchaPoints(t, s, data.ID)
	now = now.Add(captchaTTL + time.Second)
	if s.Verify(data.ID, "alice", "secret", xs, ys) {
		t.Error("an expired challenge was accepted")
	}
}

func TestVerifyFormParsesPointLists(t *testing.T) {
	s := newTestCaptchaService(time.Now)
	data, err := s.New("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	xs, ys := captchaPoints(t, s, data.ID)
	join := func(values []int) string {
		parts := make([]string, len(values))
		for i, v := range values {
			parts[i] = strconv.Itoa(v)
		}
		return strings.Join(parts, ",")
	}
	if !s.VerifyForm(data.ID, "alice", "secret", join(xs), join(ys)) {
		t.Error("VerifyForm rejected the expected comma-separated clicks")
	}

	data, err = s.New("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if s.VerifyForm(data.ID, "alice", "secret", "1,x", "1,2") {
		t.Error("VerifyForm accepted a non-numeric coordinate")
	}
	if s.VerifyForm(data.ID, "alice", "secret", "", "") {
		t.Error("VerifyForm accepted empty coordinates")
	}
}

// A challenge is tied to the credentials it was requested for: a solved one cannot be
// redirected at another account, and a wrong-credential attempt consumes it.
func TestCaptchaServiceBindsCredentials(t *testing.T) {
	s := newTestCaptchaService(time.Now)

	data, err := s.New("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	xs, ys := captchaPoints(t, s, data.ID)
	if s.Verify(data.ID, "alice", "wrong", xs, ys) {
		t.Error("a challenge issued for another password was accepted")
	}
	if s.Verify(data.ID, "alice", "secret", xs, ys) {
		t.Error("the failed credential check left the challenge replayable")
	}

	data, err = s.New("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	xs, ys = captchaPoints(t, s, data.ID)
	if s.Verify(data.ID, "bob", "secret", xs, ys) {
		t.Error("a challenge issued for another username was accepted")
	}

	data, err = s.New("bob", "other")
	if err != nil {
		t.Fatal(err)
	}
	xs, ys = captchaPoints(t, s, data.ID)
	if !s.Verify(data.ID, "bob", "other", xs, ys) {
		t.Error("the credentials the challenge was issued for were rejected")
	}
}
