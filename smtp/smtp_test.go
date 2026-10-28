package smtp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chuccp/http2smtp/model"
)

// The content type picker must not overwrite an explicit Content-Type header, and the
// check has to ignore case because HTTP header names are case-insensitive.
func TestHasHeaderIgnoresCase(t *testing.T) {
	if !hasHeader(map[string]string{"content-type": "application/xml"}, "Content-Type") {
		t.Error("lower-case content-type should be detected")
	}
	if !hasHeader(map[string]string{"CONTENT-TYPE": "application/xml"}, "Content-Type") {
		t.Error("upper-case content-type should be detected")
	}
	if hasHeader(map[string]string{"Accept": "application/json"}, "Content-Type") {
		t.Error("an unrelated header must not count as Content-Type")
	}
}

// SendAPIMail must send the configured body type, let an explicit Content-Type header
// win, and invent nothing when neither is set. OnlySendByError keeps the probe from
// trying to deliver the response by SMTP, so no mail server is needed.
func TestSendAPIMailContentType(t *testing.T) {
	var gotType, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	schedule := func() *model.Schedule {
		return &model.Schedule{
			Url: srv.URL, Method: "POST", Body: `{"a":1}`,
			ContentType: "application/json", IsOnlySendByError: true,
		}
	}

	if _, err := SendAPIMail(schedule(), &model.SMTP{}, nil); err != nil {
		t.Fatalf("send: %v", err)
	}
	if gotType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotType)
	}
	if gotBody != `{"a":1}` {
		t.Errorf("body = %q, want the configured body", gotBody)
	}

	withHeader := schedule()
	headers, err := json.Marshal([]*model.Header{{Name: "content-type", Value: "application/xml"}})
	if err != nil {
		t.Fatal(err)
	}
	withHeader.HeaderStr = string(headers)
	gotType = ""
	if _, err := SendAPIMail(withHeader, &model.SMTP{}, nil); err != nil {
		t.Fatalf("send with header: %v", err)
	}
	if gotType != "application/xml" {
		t.Errorf("Content-Type = %q, want the explicit header to win", gotType)
	}

	plain := schedule()
	plain.ContentType = ""
	gotType = "unset"
	if _, err := SendAPIMail(plain, &model.SMTP{}, nil); err != nil {
		t.Fatalf("send without type: %v", err)
	}
	if gotType != "" {
		t.Errorf("Content-Type = %q, want none when nothing is configured", gotType)
	}
}
