package util

import (
	"errors"
	"fmt"
	"regexp"
)

func FormatMail(name, mail string) string {
	if len(name) == 0 {
		return mail
	}
	return fmt.Sprintf(`"%s" <%s>`, name, mail)
}

func ParseMail(recipient string) (name string, mail string, err error) {
	str := `"(.+)" <(.+@.+)>`
	re := regexp.MustCompile(str)
	match := re.FindStringSubmatch(recipient)
	if len(match) == 3 {
		return match[1], match[2], nil
	}
	str = `(.+) <(.+@.+)>`
	re = regexp.MustCompile(str)
	match = re.FindStringSubmatch(recipient)
	if len(match) == 3 {
		return match[1], match[2], nil
	}
	str = `([a-zA-Z0-9]+@[a-zA-Z0-9\.]+)`
	re = regexp.MustCompile(str)
	match = re.FindStringSubmatch(recipient)
	if len(match) == 2 {
		return "", match[0], nil
	}
	return "", "", errors.New("recipient Format error")

}

func ExtractEmails(text string) []string {
	// 定义正则表达式
	emailRegex := regexp.MustCompile(
		`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)

	// 使用FindAllString方法查找所有匹配的电子邮件地址
	emails := emailRegex.FindAllString(text, -1)

	// 去除重复的电子邮件地址
	uniqueEmails := removeDuplicates(emails)

	return uniqueEmails
}

// removeDuplicates 函数用于去除电子邮件地址列表中的重复项
func removeDuplicates(emails []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range emails {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
