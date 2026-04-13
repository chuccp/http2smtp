package smtp

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/template"
)

func TestSendAPIMail(t *testing.T) {
	bodyString := `{"name":"aaaa"}`
	parse, err := template.New("template").Parse("{{.name}}")
	if err == nil {
		buffer := new(bytes.Buffer)
		data := make(map[string]interface{})
		err = json.Unmarshal([]byte(bodyString), &data)
		if err == nil {
			err = parse.Execute(buffer, data)
			if err == nil {
				println(buffer.String())
			}
		}

	}

}
