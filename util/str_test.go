package util

import (
	"reflect"
	"testing"
)

func TestSplitAndDeduplicate(t *testing.T) {
	got := SplitAndDeduplicate(" a, b ,a,,c ", ",")
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("SplitAndDeduplicate = %#v, want %#v", got, want)
	}
	if got := SplitAndDeduplicate("", ","); len(got) != 0 {
		t.Errorf("SplitAndDeduplicate(\"\") = %#v, want empty", got)
	}
}

func TestDeduplicateIds(t *testing.T) {
	if got := DeduplicateIds("1,2,1,3"); got != "1,2,3" {
		t.Errorf("DeduplicateIds = %q, want 1,2,3", got)
	}
	if got := DeduplicateIds(""); got != "" {
		t.Errorf("DeduplicateIds(\"\") = %q, want empty", got)
	}
}

func TestStringToUintIds(t *testing.T) {
	got := StringToUintIds("1,2,x,,3")
	if want := []uint{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("StringToUintIds = %#v, want %#v", got, want)
	}
	if got := StringToUintIds(""); len(got) != 0 {
		t.Errorf("StringToUintIds(\"\") = %#v, want empty", got)
	}
}

func TestContainsAnyIgnoreCase(t *testing.T) {
	if !ContainsAnyIgnoreCase("Application/JSON; charset=utf-8", "json") {
		t.Error("case-insensitive containment failed")
	}
	if ContainsAnyIgnoreCase("text/plain", "json", "xml") {
		t.Error("unrelated content type matched")
	}
}
