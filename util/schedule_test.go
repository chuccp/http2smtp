package util

import "testing"

func TestParserCron(t *testing.T) {
	err := ParserCron("0 0 0/1 * * ?")
	if err != nil {
		t.Error(err)
	}
}
