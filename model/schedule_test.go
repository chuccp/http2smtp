package model

import "testing"

// The runner reloads a cron entry only when its Key changes, so every field the request
// is built from must take part in the hash — the content type was easy to miss.
func TestScheduleKeyIncludesContentType(t *testing.T) {
	base := func() *Schedule {
		return &Schedule{
			UserId: 1, Name: "n", TokenId: 2, Cron: "0 0 9 * * *", Url: "http://example.com",
			Method: "POST", Body: "b", ContentType: "application/json", Template: "t", IsUse: true,
		}
	}
	a, b := base(), base()
	if a.Key() != b.Key() {
		t.Fatal("identical schedules should hash the same")
	}
	b.ContentType = "text/plain"
	if a.Key() == b.Key() {
		t.Error("Key must change when the content type changes")
	}
}
