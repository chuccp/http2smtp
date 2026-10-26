package service

import (
	"path/filepath"
	"testing"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/http2smtp/model"
)

// newTestUserService wires a UserService against a throwaway sqlite file, with no
// core.Context — UpdateUser only touches the user model.
func newTestUserService(t *testing.T) *UserService {
	t.Helper()
	conn, err := db.ConnectionSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening sqlite: %v", err)
	}
	// Close the underlying handle so the temp file can be removed on Windows
	sqlDB, err := conn.GetGorm().DB()
	if err != nil {
		t.Fatalf("getting sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	userModel := &model.UserModel{}
	if err := userModel.Init(conn, nil); err != nil {
		t.Fatalf("init user model: %v", err)
	}
	if err := userModel.CreateTable(); err != nil {
		t.Fatalf("create t_user: %v", err)
	}
	return &UserService{userModel: userModel}
}

func mustCreateUser(t *testing.T, s *UserService) uint {
	t.Helper()
	if err := s.CreateUser("alice", "password123", false, true); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	user, err := s.userModel.FindOneByName("alice")
	if err != nil || user == nil {
		t.Fatalf("reading back the created user: %v", err)
	}
	return user.Id
}

// Disabling an account must also end the sessions it already had open: auth.User only
// checks the salt, so an unchanged salt would leave the disabled user's cookie working.
func TestUpdateUser_DisableRotatesSalt(t *testing.T) {
	s := newTestUserService(t)
	id := mustCreateUser(t, s)

	before, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK: %v", err)
	}
	if err := s.UpdateUser(id, "alice", "", false, false); err != nil {
		t.Fatalf("UpdateUser(disable): %v", err)
	}
	after, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK after: %v", err)
	}
	if after.Salt == before.Salt {
		t.Error("disabling a user left the session salt unchanged; existing sessions would keep working")
	}
	if after.IsUse {
		t.Error("IsUse was not persisted as false")
	}
}

// An update that leaves the account enabled is not a revocation and must not end
// the user's sessions.
func TestUpdateUser_KeepEnabledKeepsSalt(t *testing.T) {
	s := newTestUserService(t)
	id := mustCreateUser(t, s)

	before, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK: %v", err)
	}
	if err := s.UpdateUser(id, "alice", "", false, true); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	after, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK after: %v", err)
	}
	if after.Salt != before.Salt {
		t.Error("an update that kept the user enabled rotated the salt and ended all sessions")
	}
}

// Re-enabling a user does not need to revoke anything, but must persist.
func TestUpdateUser_ReEnablePersists(t *testing.T) {
	s := newTestUserService(t)
	id := mustCreateUser(t, s)

	if err := s.UpdateUser(id, "alice", "", false, false); err != nil {
		t.Fatalf("UpdateUser(disable): %v", err)
	}
	if err := s.UpdateUser(id, "alice", "", false, true); err != nil {
		t.Fatalf("UpdateUser(re-enable): %v", err)
	}
	after, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK: %v", err)
	}
	if !after.IsUse {
		t.Error("IsUse was not persisted as true")
	}
}

// A password change rotates the salt by design — it ends every open session.
func TestUpdateUser_PasswordChangeRotatesSalt(t *testing.T) {
	s := newTestUserService(t)
	id := mustCreateUser(t, s)

	before, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK: %v", err)
	}
	if err := s.UpdateUser(id, "alice", "newpassword1", false, true); err != nil {
		t.Fatalf("UpdateUser(password): %v", err)
	}
	after, err := s.userModel.FindByPK(id)
	if err != nil {
		t.Fatalf("FindByPK after: %v", err)
	}
	if after.Salt == before.Salt {
		t.Error("a password change left the session salt unchanged")
	}
}
