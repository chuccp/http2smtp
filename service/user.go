package service

import (
	"errors"
	"time"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	localutil "github.com/chuccp/http2smtp/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// newSessionSalt returns a fresh session epoch for a user.
//
// Salt is not part of password hashing — bcrypt carries its own — it exists solely so a
// previously issued session can be invalidated: the token embeds the salt it was minted
// with, and User() only accepts it while it still matches the stored value.
func newSessionSalt() string {
	return uuid.NewString()
}

type UserService struct {
	context   *core.Context
	userModel *model.UserModel
}

func (s *UserService) Init(context *core.Context) error {
	s.context = context
	s.userModel = wf.GetModel[*model.UserModel](context)
	return nil
}

// dbReady reports whether the database has been attached. Sign-in and the setup probes
// are reachable before the wizard has run, where a query would dereference the model's
// nil connection instead of answering.
func (s *UserService) dbReady() bool {
	return model.CoreState(s.context.GetConfig()).DbInit
}

// ValidateLogin verifies username and password against the database.
// Returns the user record if successful, or an error if authentication fails.
func (s *UserService) ValidateLogin(username, password string) (*model.User, error) {
	if !s.dbReady() {
		return nil, errors.New("database is not initialized, please complete setup first")
	}
	user, err := s.userModel.FindOneByName(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}
	if !user.IsUse {
		return nil, errors.New("user is disabled")
	}
	if !localutil.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

// GetPage returns a paginated list of users.
func (s *UserService) GetPage(page *web.Page) (any, error) {
	return s.userModel.Query().Order("id desc").PageForWeb(page)
}

// GetOne returns a user by ID.
func (s *UserService) GetOne(id uint) (*model.User, error) {
	user, err := s.userModel.FindByPK(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// CreateUser creates a new user with a hashed password.
func (s *UserService) CreateUser(name, password string, isAdmin, isUse bool) error {
	existing, err := s.userModel.FindOneByName(name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return errors.New("username already exists")
	}
	hashedPassword, err := localutil.HashPassword(password)
	if err != nil {
		return err
	}
	user := &model.User{
		Name:       name,
		Password:   hashedPassword,
		IsAdmin:    isAdmin,
		IsUse:      isUse,
		Salt:       newSessionSalt(),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	return s.userModel.Save(user)
}

// UpdateUser updates an existing user. If password is non-empty, it will be hashed and updated.
func (s *UserService) UpdateUser(id uint, name string, password string, isAdmin, isUse bool) error {
	user, err := s.userModel.FindByPK(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if new name conflicts with existing users (excluding self)
	if name != user.Name {
		existing, err := s.userModel.FindOneByName(name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if existing != nil {
			return errors.New("username already exists")
		}
		user.Name = name
	}

	if password != "" {
		hashedPassword, err := localutil.HashPassword(password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
		// A password change ends every session this user had open
		user.Salt = newSessionSalt()
	}

	user.IsAdmin = isAdmin
	// Disabling an account must also end the sessions it already had open: User() only
	// checks the salt, never is_use, so rotate it here the same way DeleteUser does.
	if user.IsUse && !isUse {
		user.Salt = newSessionSalt()
	}
	user.IsUse = isUse
	user.UpdateTime = time.Now()

	// Save, not UpdateByPK: the update path goes through gorm's Updates(struct), which
	// drops zero values — a false IsAdmin/IsUse would be silently ignored and the
	// demote/disable would look like it succeeded. Save writes every column.
	return s.userModel.Save(user)
}

// ChangeOwnPassword replaces the signed-in user's password. The salt rotates with it, so
// every session that user had open ends — including the one that asked for the change.
func (s *UserService) ChangeOwnPassword(id uint, password string) error {
	user, err := s.userModel.FindByPK(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	hashed, err := localutil.HashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hashed
	user.Salt = newSessionSalt()
	user.UpdateTime = time.Now()
	return s.userModel.Save(user)
}

// HasAdminUser checks if any admin user exists in the database.
func (s *UserService) HasAdminUser() (bool, error) {
	if !s.dbReady() {
		return false, errors.New("database is not initialized")
	}
	count, err := s.userModel.Query().Where("is_admin = ?", true).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAdminUsername returns the username of the first admin user.
func (s *UserService) GetAdminUsername() (string, error) {
	user, err := s.userModel.Query().Where("is_admin = ?", true).One()
	if err != nil || user == nil {
		return "", err
	}
	return user.Name, nil
}

// CreateAdminUser creates an admin user during system initialization.
func (s *UserService) CreateAdminUser(name, password string) error {
	existing, err := s.userModel.FindOneByName(name)

	if errors.Is(err, gorm.ErrRecordNotFound) || existing == nil {
		hashedPassword, err := localutil.HashPassword(password)
		if err != nil {
			return err
		}
		user := &model.User{
			Name:       name,
			Password:   hashedPassword,
			IsAdmin:    true,
			IsUse:      true,
			Salt:       newSessionSalt(),
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		return s.userModel.Save(user)
	}
	return errors.New("admin is exist")

}

// ResetAdminPassword resets the password of the existing admin user during setup.
func (s *UserService) ResetAdminPassword(username, password string) error {
	user, err := s.userModel.Query().Where("is_admin = ?", true).One()
	if err != nil || user == nil {
		return errors.New("admin user not found")
	}
	hashedPassword, err := localutil.HashPassword(password)
	if err != nil {
		return err
	}
	user.Name = username
	user.Password = hashedPassword
	user.Salt = newSessionSalt()
	user.UpdateTime = time.Now()
	return s.userModel.UpdateByPK(user)
}

// RotateSalt invalidates every session already issued for the user, including the
// caller's own. Used on sign-out so a captured token stops working immediately.
func (s *UserService) RotateSalt(id uint) error {
	user, err := s.userModel.FindByPK(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	user.Salt = newSessionSalt()
	user.UpdateTime = time.Now()
	return s.userModel.UpdateByPK(user)
}

// DeleteUser soft-deletes a user by setting is_use = false.
func (s *UserService) DeleteUser(id uint) error {
	user, err := s.userModel.FindByPK(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	user.IsUse = false
	// User() only checks the salt, not is_use, so rotate it here too: disabling an
	// account must also end the sessions it already had open.
	user.Salt = newSessionSalt()
	user.UpdateTime = time.Now()
	// Save for the same reason as UpdateUser: a struct update would skip the false
	// IsUse and leave the account able to sign in.
	return s.userModel.Save(user)
}

// FillUserNames fills the UserName field for a list of records that have UserId.
func (s *UserService) FillUserNames(userIds []uint, setUserName func(userId uint, userName string)) {
	if len(userIds) == 0 {
		return
	}
	nameMap, err := s.userModel.FindNameMapByIds(userIds)
	if err != nil {
		return
	}
	for _, uid := range userIds {
		if name, ok := nameMap[uid]; ok {
			setUserName(uid, name)
		}
	}
}

// AdminIds reports which of userIds belong to administrators. The list pages use it to
// mark admin-created rows for an admin viewer.
func (s *UserService) AdminIds(userIds []uint) map[uint]bool {
	admins := make(map[uint]bool)
	if len(userIds) == 0 {
		return admins
	}
	users, err := s.userModel.Query().Where("id IN ? AND is_admin = ?", userIds, true).All()
	if err != nil {
		return admins
	}
	for _, u := range users {
		admins[u.Id] = true
	}
	return admins
}
