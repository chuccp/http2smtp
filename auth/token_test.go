package auth

import (
	"testing"

	"github.com/chuccp/go-web-frame/util"
)

func TestGetShareCode(t *testing.T) {

	t.Log(util.Second())
	t.Log(util.Second() % 1000000000)
	t.Log(GetShareCode())
}
