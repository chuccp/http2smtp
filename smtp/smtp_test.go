package smtp

import "testing"

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
