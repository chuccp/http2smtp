package auth

import (
	"crypto/rand"
	"encoding/json"

	"github.com/chuccp/go-web-frame/util"
)

// cryptoRandomAlnum fills n characters from [A-Za-z0-9] using the cryptographic
// generator. Session cookies are only as strong as the key that encrypts them, so
// this must not be the time-seeded math/rand the framework helpers use.
func cryptoRandomAlnum(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out := make([]byte, n)
	buf := make([]byte, n)
	for i := 0; i < n; {
		if _, err := rand.Read(buf); err != nil {
			// crypto/rand.Read only fails if the OS source is broken; a session is
			// not worth silently weakening the key for.
			panic("crypto/rand unavailable: " + err.Error())
		}
		for _, b := range buf {
			// Rejection sampling keeps the distribution uniform: 248 is the
			// largest multiple of 62 below 256.
			if b >= 248 {
				continue
			}
			out[i] = charset[int(b)%len(charset)]
			i++
			if i == n {
				break
			}
		}
	}
	return string(out)
}

// The key and IV are minted per process, so a restart ends every outstanding session —
// the same behaviour as rotating the salt, just coarser.
var aesKey = cryptoRandomAlnum(32)
var aesIv = cryptoRandomAlnum(16)

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
