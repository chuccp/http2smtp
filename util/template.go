package util

import (
	"bytes"
	"encoding/json"
	"text/template"
)

func ParseTemplate(templateStr string, bodyString string) (string, error) {
	parse, err := template.New("template").Parse(templateStr)
	if err == nil {
		buffer := new(bytes.Buffer)
		data := make(map[string]interface{})
		err = json.Unmarshal([]byte(bodyString), &data)
		if err == nil {
			err = parse.Execute(buffer, data)
			if err == nil {
				return buffer.String(), nil
			}
		}
	}
	return "", err
}
