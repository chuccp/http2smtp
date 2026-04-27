package service

import (
	"errors"
	"time"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
	localutil "github.com/chuccp/http2smtp/util"
	"gorm.io/gorm"
)

type UserService struct {
	context   *core.Context
	userModel *model.UserModel
}

func (s *UserService) Init(context *core.Context) error {
	s.context = context
	s.userModel = wf.GetModel[*model.UserModel](context)
	return nil
}

// ValidateLogin verifies username and password against the database.
// Returns the user record if successful, or an error if authentication fails.
func (s *UserService) ValidateLogin(username, password string) (*model.User, error) {
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
	user, err := s.userModel.FindById(id)
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
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	return s.userModel.Save(user)
}

// UpdateUser updates an existing user. If password is non-empty, it will be hashed and updated.
func (s *UserService) UpdateUser(id uint, name string, password string, isAdmin, isUse bool) error {
	user, err := s.userModel.FindById(id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if new name conflicts with existing users (excluding self)
	if name != user.Name {
		existing, err := s.userModel.FindOneByName(name)
		if err != nil {
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
	}

	user.IsAdmin = isAdmin
	user.IsUse = isUse
	user.UpdateTime = time.Now()

	return s.userModel.UpdateById(user)
}

// HasAdminUser checks if any admin user exists in the database.
func (s *UserService) HasAdminUser() (bool, error) {
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
	user.UpdateTime = time.Now()
	return s.userModel.UpdateById(user)
}

// DeleteUser soft-deletes a user by setting is_use = false.
func (s *UserService) DeleteUser(id uint) error {
	user, err := s.userModel.FindById(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	user.IsUse = false
	user.UpdateTime = time.Now()
	return s.userModel.UpdateById(user)
}
