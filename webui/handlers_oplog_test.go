package webui

import (
	"testing"

	"github.com/chuccp/http2smtp/entity"
)

// Every action and target an entry can carry must have a label and a tag class; adding a
// constant without wiring it here leaves the page showing raw values.
func TestOplogLabelsCoverEveryActionAndTarget(t *testing.T) {
	tr := translatorFor("zh-cn")

	actions := []string{
		entity.OpCreate, entity.OpUpdate, entity.OpDelete,
		entity.OpEnable, entity.OpDisable, entity.OpTrigger,
		entity.OpCleanup, entity.OpRestart,
		entity.OpLogin, entity.OpLoginFailed, entity.OpLogout,
	}
	for _, action := range actions {
		key, ok := oplogActionKeys[action]
		if !ok {
			t.Errorf("action %q has no translation key", action)
			continue
		}
		if got := tr.T(key); got == key {
			t.Errorf("action %q key %q is missing from the locale", action, key)
		}
		if _, ok := oplogActionClasses[action]; !ok {
			t.Errorf("action %q has no tag class", action)
		}
	}

	targets := []string{
		entity.TargetSMTP, entity.TargetMail, entity.TargetToken, entity.TargetSchedule,
		entity.TargetUser, entity.TargetLog, entity.TargetSettings, entity.TargetOplog,
	}
	for _, target := range targets {
		key, ok := oplogTargetKeys[target]
		if !ok {
			t.Errorf("target %q has no translation key", target)
			continue
		}
		if got := tr.T(key); got == key {
			t.Errorf("target %q key %q is missing from the locale", target, key)
		}
	}
}
