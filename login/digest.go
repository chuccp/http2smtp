package login

import (
	"errors"
	"github.com/chuccp/http2smtp/util"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type client struct {
	isValid   bool
	last_seen int
}

type digest_client struct {
	maxSize int
	timeOut int
	data    []string
	dataMap map[string]*client
	off     int
	lock    *sync.Mutex
}

func newDigestClient() *digest_client {
	dc := &digest_client{maxSize: 1000, timeOut: 3600 * 24}
	dc.data = make([]string, 1000)
	dc.dataMap = make(map[string]*client)
	dc.lock = new(sync.Mutex)
	dc.off = 0
	return dc
}
func (dc *digest_client) hasClient(key string) bool {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	v, ok := dc.dataMap[key]
	if !ok {
		return false
	} else {
		return time.Now().Second()-v.last_seen < dc.timeOut
	}
}

func (dc *digest_client) isValid(key string) bool {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	v, ok := dc.dataMap[key]
	if !ok {
		return false
	} else {
		if !v.isValid {
			return false
		}
		return time.Now().Second()-v.last_seen < dc.timeOut
	}
}

func (dc *digest_client) toValid(key string) bool {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	v, ok := dc.dataMap[key]
	if !ok {
		return false
	} else {
		v.isValid = true
		v.last_seen = time.Now().Second()
		return true
	}
}

func (dc *digest_client) deleteClient(key string) {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	delete(dc.dataMap, key)
}
func (dc *digest_client) getNew() string {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	if dc.off >= dc.maxSize {
		dc.off = 0
	}
	key := RandomKey()
	v := dc.data[dc.off]
	if len(v) > 0 {
		delete(dc.dataMap, v)
	}
	dc.data[dc.off] = key
	dc.dataMap[key] = &client{last_seen: time.Now().Second(), isValid: false}
	dc.off++
	return key
}

// SecretProvider key =  md5(md5(p)+u)
// sign = md5(key+nonce)
type SecretProvider func(user string) string

type DigestAuth struct {
	secretProvider SecretProvider
	digestClient   *digest_client
}

func (digestAuth *DigestAuth) getNonce(ctx *gin.Context) string {
	nonce := ctx.GetHeader("Nonce")
	if len(nonce) == 0 {
		cookie, err := ctx.Cookie("Nonce")
		if err == nil {
			nonce = cookie
		}
	}
	return nonce
}
func (digestAuth *DigestAuth) JustCheck(ctx *gin.Context) (any, error) {
	nonce := digestAuth.getNonce(ctx)
	val := digestAuth.digestClient.isValid(nonce)
	if val {
		return nil, nil
	}
	ctx.Status(http.StatusUnauthorized)
	return "login timeout", errors.New("login timeout, please refresh the page")
}

func (digestAuth *DigestAuth) HasSign(ctx *gin.Context) bool {
	nonce := digestAuth.getNonce(ctx)
	return digestAuth.digestClient.isValid(nonce)
}
func (digestAuth *DigestAuth) Logout(ctx *gin.Context) (any, error) {
	nonce := digestAuth.getNonce(ctx)
	digestAuth.digestClient.deleteClient(nonce)
	return "success", nil
}
func (digestAuth *DigestAuth) CheckSign(ctx *gin.Context) (any, error) {
	if strings.EqualFold(ctx.Request.Method, "get") {
		key := digestAuth.digestClient.getNew()
		var authInfo AuthInfo
		authInfo.Nonce = key
		return &authInfo, nil
	} else {
		var u User
		err := ctx.ShouldBindBodyWithJSON(&u)
		if err != nil {
			return nil, err
		}
		if digestAuth.digestClient.hasClient(u.Nonce) {
			key := digestAuth.secretProvider(u.Username)
			sign := util.MD5Str(key + u.Nonce)
			if strings.EqualFold(sign, u.Response) {
				digestAuth.digestClient.toValid(u.Nonce)

				host, _, _ := net.SplitHostPort(ctx.Request.Host)
				ctx.SetSameSite(http.SameSiteNoneMode)
				ctx.SetCookie("Nonce", u.Nonce, 3600*24*30, "/", host, true, true)
				return "success", nil
			}
		}

	}
	ctx.Status(http.StatusUnauthorized)
	return "username or password is incorrect", nil
}

func NewDigestAuth(secretProvider SecretProvider) *DigestAuth {
	return &DigestAuth{secretProvider: secretProvider, digestClient: newDigestClient()}
}
