package rest

import (
	"testing"

	"github.com/chuccp/http2smtp/entity"
)

// resolveState decides the state a caller may set on a token. The sensitive case is
// re-enabling: "admin disabled" is the administrator's disposal and must survive a
// plain owner asking for state = 0.
func TestResolveState(t *testing.T) {
	token := &Token{}

	t.Run("owner cannot re-enable an admin-disabled token", func(t *testing.T) {
		got := token.resolveState(entity.TokenStateInUse, entity.TokenStateAdminDisabled, false)
		if got != entity.TokenStateAdminDisabled {
			t.Errorf("owner re-enable of an admin-disabled token: got %d, want %d", got, entity.TokenStateAdminDisabled)
		}
	})

	t.Run("admin can re-enable an admin-disabled token", func(t *testing.T) {
		got := token.resolveState(entity.TokenStateInUse, entity.TokenStateAdminDisabled, true)
		if got != entity.TokenStateInUse {
			t.Errorf("admin re-enable: got %d, want %d", got, entity.TokenStateInUse)
		}
	})

	t.Run("owner can re-enable their own user-disabled token", func(t *testing.T) {
		got := token.resolveState(entity.TokenStateInUse, entity.TokenStateUserDisabled, false)
		if got != entity.TokenStateInUse {
			t.Errorf("owner re-enable of a user-disabled token: got %d, want %d", got, entity.TokenStateInUse)
		}
	})

	t.Run("a fresh create always starts in use", func(t *testing.T) {
		got := token.resolveState(entity.TokenStateInUse, entity.TokenStateInUse, false)
		if got != entity.TokenStateInUse {
			t.Errorf("create: got %d, want %d", got, entity.TokenStateInUse)
		}
	})

	t.Run("a disable request is stamped by role", func(t *testing.T) {
		if got := token.resolveState(entity.TokenStateUserDisabled, entity.TokenStateInUse, true); got != entity.TokenStateAdminDisabled {
			t.Errorf("admin disable: got %d, want %d", got, entity.TokenStateAdminDisabled)
		}
		if got := token.resolveState(entity.TokenStateUserDisabled, entity.TokenStateInUse, false); got != entity.TokenStateUserDisabled {
			t.Errorf("user disable: got %d, want %d", got, entity.TokenStateUserDisabled)
		}
	})
}
