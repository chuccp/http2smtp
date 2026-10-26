package rest

import (
	"testing"

	"github.com/chuccp/go-web-frame/web"
)

func page(no, size int) *web.Page { return &web.Page{PageNo: no, PageSize: size} }

func TestCapPage(t *testing.T) {
	t.Run("oversized page is clamped", func(t *testing.T) {
		got := capPage(page(1, 1000000))
		if got.PageSize != maxPageSize {
			t.Errorf("pageSize = %d, want %d", got.PageSize, maxPageSize)
		}
	})
	t.Run("sane page is untouched", func(t *testing.T) {
		got := capPage(page(3, 50))
		if got.PageNo != 3 || got.PageSize != 50 {
			t.Errorf("page = %+v, want {3 50}", got)
		}
	})
	t.Run("zero values get defaults", func(t *testing.T) {
		got := capPage(page(0, 0))
		if got.PageNo != 1 || got.PageSize != 10 {
			t.Errorf("page = %+v, want {1 10}", got)
		}
	})
	t.Run("nil is safe", func(t *testing.T) {
		got := capPage(nil)
		if got.PageNo != 1 || got.PageSize != 10 {
			t.Errorf("page = %+v, want {1 10}", got)
		}
	})
}
