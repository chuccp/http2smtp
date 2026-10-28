package service

import (
	"testing"

	"github.com/chuccp/http2smtp/model"
)

// The token's configured list is the default, the request adds to it, and an address in
// both lists is sent to once — case-insensitively.
func TestMergeRecipients(t *testing.T) {
	a := &model.Mail{Name: "A", Mail: "a@example.com"}
	b := &model.Mail{Name: "B", Mail: "b@example.com"}

	merged := MergeRecipients([]*model.Mail{a, nil, b}, []string{" B@example.com ", "c@example.com", "", "a@example.com"})
	if len(merged) != 3 {
		t.Fatalf("merged = %+v, want 3 recipients", merged)
	}
	if merged[0] != a || merged[1] != b {
		t.Error("configured rows should be kept as they are")
	}
	if merged[2].Mail != "c@example.com" || merged[2].Name != "" {
		t.Errorf("requested recipient = %+v", merged[2])
	}

	only := MergeRecipients(nil, []string{"x@example.com", "X@example.com", " "})
	if len(only) != 1 || only[0].Mail != "x@example.com" {
		t.Errorf("request-only merge = %+v, want one x@example.com", only)
	}

	if got := MergeRecipients([]*model.Mail{{Name: "A", Mail: "a@example.com"}}, nil); len(got) != 1 {
		t.Errorf("configured-only merge = %+v, want the configured row", got)
	}
}
