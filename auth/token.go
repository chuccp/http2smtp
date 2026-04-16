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

var aesKey = util.GenerateRandomStringByAlphanumeric(32)
var aesIv = util.GenerateRandomStringByAlphanumeric(16)

func GenerateToken(value string, code string, type_ string) (string, error) {
	token := &Token{
		Time:  util.NowDateTime(),
		Type:  type_,
		Value: value,
		Code:  code,
	}
	v, _ := json.Marshal(token)
	return util.EncryptByCBC(string(v), aesKey, aesIv)
}
func Encrypt(value any) (string, error) {
	v, _ := json.Marshal(value)
	return util.EncryptByCBC(string(v), aesKey, aesIv)
}

func Decrypt(token string, v any) error {
	str, err := util.DecryptByCBC(token, aesKey, aesIv)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(str), v)
	return err
}
