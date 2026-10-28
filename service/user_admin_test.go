package service

import "testing"

// AdminIds is what the list pages use to star admin-created rows; it must report only
// the ids that really belong to administrators.
func TestUserServiceAdminIds(t *testing.T) {
	s := newTestUserService(t)
	if err := s.CreateUser("admin", "password123", true, true); err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if err := s.CreateUser("bob", "password123", false, true); err != nil {
		t.Fatalf("create bob: %v", err)
	}
	admin, err := s.userModel.FindOneByName("admin")
	if err != nil || admin == nil {
		t.Fatalf("load admin: %v", err)
	}
	bob, err := s.userModel.FindOneByName("bob")
	if err != nil || bob == nil {
		t.Fatalf("load bob: %v", err)
	}

	admins := s.AdminIds([]uint{admin.Id, bob.Id, 999})
	if !admins[admin.Id] {
		t.Error("the administrator should be reported as an admin")
	}
	if admins[bob.Id] {
		t.Error("a regular user must not be reported as an admin")
	}
	if admins[999] {
		t.Error("an unknown id must not be reported as an admin")
	}
	if got := s.AdminIds(nil); len(got) != 0 {
		t.Errorf("AdminIds(nil) = %v, want empty", got)
	}
}
