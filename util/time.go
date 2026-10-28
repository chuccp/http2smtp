package util

import (
	"time"
)

var DateFormat = "2006-01-02"

func FormatDate(tm time.Time) string {
	return tm.Format(DateFormat)
}
