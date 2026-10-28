package model

import (
	"encoding/json"
	"testing"

	"github.com/chuccp/go-web-frame/config"
)

func TestParseByteSize(t *testing.T) {
	cases := []struct {
		in   string
		want ByteSize
	}{
		{"10M", 10 << 20},
		{"10MB", 10 << 20},
		{"10mb", 10 << 20},
		{" 10 M ", 10 << 20},
		{"512K", 512 << 10},
		{"1G", 1 << 30},
		{"10485760", 10 << 20},
		{"10B", 10},
		{"-1", -1},
	}
	for _, c := range cases {
		got, err := ParseByteSize(c.in)
		if err != nil {
			t.Errorf("ParseByteSize(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseByteSize(%q) = %d, want %d", c.in, got, c.want)
		}
	}

	for _, bad := range []string{"", "0", "-2", "M", "abc", "1.5M"} {
		if _, err := ParseByteSize(bad); err == nil {
			t.Errorf("ParseByteSize(%q) should fail", bad)
		}
	}
}

func TestByteSizeString(t *testing.T) {
	cases := map[ByteSize]string{
		10 << 20:  "10M",
		1 << 30:   "1G",
		512 << 10: "512K",
		10:        "10",
		-1:        "-1",
	}
	for in, want := range cases {
		if got := in.String(); got != want {
			t.Errorf("ByteSize(%d).String() = %q, want %q", in, got, want)
		}
	}
}

func TestByteSizeJSON(t *testing.T) {
	for _, raw := range []string{`"10M"`, `10485760`} {
		var b ByteSize
		if err := json.Unmarshal([]byte(raw), &b); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		if b != 10<<20 {
			t.Errorf("unmarshal %s = %d, want %d", raw, b, 10<<20)
		}
	}
	var b ByteSize
	if err := json.Unmarshal([]byte(`0`), &b); err == nil {
		t.Error("a zero size should be rejected")
	}
	out, err := json.Marshal(ByteSize(10 << 20))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "10485760" {
		t.Errorf("marshal = %s, want the byte count", out)
	}
}

func TestConfigReadsMaxBodySize(t *testing.T) {
	raw := []byte("[core]\nmaxbodysize=10M\n")
	c, err := config.NewFromBytes(raw, "ini")
	if err != nil {
		t.Fatalf("loading ini: %v", err)
	}
	cfg, err := GetConfig(c)
	if err != nil {
		t.Fatalf("decoding config: %v", err)
	}
	if cfg.Core.MaxBodySize != 10<<20 {
		t.Errorf("maxBodySize = %d, want %d", cfg.Core.MaxBodySize, 10<<20)
	}
	if got := cfg.Core.MaxBodySizeValue(); got != 10<<20 {
		t.Errorf("MaxBodySizeValue = %d, want %d", got, 10<<20)
	}

	// An omitted key keeps the 10 MB default; -1 means unlimited.
	c, err = config.NewFromBytes([]byte("[core]\ninit=true\n"), "ini")
	if err != nil {
		t.Fatalf("loading ini: %v", err)
	}
	cfg, err = GetConfig(c)
	if err != nil {
		t.Fatalf("decoding config: %v", err)
	}
	if cfg.Core.MaxBodySize != ByteSize(DefaultMaxBodySize) {
		t.Errorf("default maxBodySize = %d, want %d", cfg.Core.MaxBodySize, DefaultMaxBodySize)
	}

	c, err = config.NewFromBytes([]byte("[core]\nmaxbodysize=-1\n"), "ini")
	if err != nil {
		t.Fatalf("loading ini: %v", err)
	}
	cfg, err = GetConfig(c)
	if err != nil {
		t.Fatalf("decoding config: %v", err)
	}
	if got := cfg.Core.MaxBodySizeValue(); got != -1 {
		t.Errorf("unlimited MaxBodySizeValue = %d, want -1", got)
	}
}

func TestCoreConfigMaxBodySizeValueFallback(t *testing.T) {
	var nilCore *CoreConfig
	if got := nilCore.MaxBodySizeValue(); got != DefaultMaxBodySize {
		t.Errorf("nil core = %d, want %d", got, DefaultMaxBodySize)
	}
	if got := (&CoreConfig{}).MaxBodySizeValue(); got != DefaultMaxBodySize {
		t.Errorf("zero core = %d, want %d", got, DefaultMaxBodySize)
	}
}
