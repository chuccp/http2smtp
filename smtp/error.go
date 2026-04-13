package smtp

import "strings"

func ToUserNotFoundError(mails []string) *UserNotFoundError {
	return &UserNotFoundError{mails: mails}
}

type UserNotFoundError struct {
	mails []string
}

func (e *UserNotFoundError) Error() string {
	return "User not found: " + strings.Join(e.mails, ",")
}
