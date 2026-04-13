package util

import (
	"regexp"
	"testing"
)

func TestName(t *testing.T) {
	re := regexp.MustCompile("^/assets")
	fa := re.MatchString("/assets/log-BwvnxVjg.js")
	t.Log(fa)
}

func TestFormatMail(t *testing.T) {

	str := `([a-zA-Z0-9]+@[a-zA-Z0-9\.]+)`

	re := regexp.MustCompile(str)
	match := re.FindStringSubmatch(`chuccp@163.com`)
	t.Log(match)
	if len(match) == 3 {
		name := match[1]
		email := match[2]
		t.Log(name, email)
	}

}
func TestFormatMail2(t *testing.T) {
	strlist := []string{"chuccp@163.com", "chuccp@163.com"}
	strlist = nil
	println(len(strlist))
}
