package auth

import (
	"encoding/json"

	"github.com/chuccp/go-web-frame/util"
)

type Token struct {
	Time  string `json:"time"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Code  string `json:"code"`
}

var aes_key = util.GenerateRandomStringByAlphanumeric(32)
var aes_iv = util.GenerateRandomStringByAlphanumeric(16)

func GenerateToken(value string, code string, type_ string) string {
	token := &Token{
		Time:  util.NowDateTime(),
		Type:  type_,
		Value: value,
		Code:  code,
	}
	v, _ := json.Marshal(token)
	return util.EncryptByCBC(string(v), aes_key, aes_iv)
}
func Encrypt(value any) string {
	v, _ := json.Marshal(value)
	return util.EncryptByCBC(string(v), aes_key, aes_iv)
}

func Decrypt(token string, v any) error {
	str, err := util.DecryptByCBC(token, aes_key, aes_iv)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(str), v)
	return err
}
