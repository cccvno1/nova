package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/cccvno1/nova/internal/model"
	"github.com/cccvno1/nova/internal/repository"
	"github.com/cccvno1/nova/pkg/auth"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
	jwtAuth  *auth.JWTAuth
}

func NewUserService(db *gorm.DB, jwtAuth *auth.JWTAuth) *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(db, true), // true 表示启用缓存
		jwtAuth:  jwtAuth,
	}
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=32"`
	Nickname string `json:"nickname" validate:"omitempty,max=50"`
}

type UpdateUserRequest struct {
	Nickname string `json:"nickname" validate:"omitempty,max=50"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}

func (s *UserService) Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}
	if exists {
		return nil, errors.New(errors.ErrRecordExists, "username already exists")
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}
	if exists {
		return nil, errors.New(errors.ErrRecordExists, "email already exists")
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashPassword(req.Password),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	return s.toResponse(user), nil
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrRecordNotFound, "user not found")
		}
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	return s.toResponse(user), nil
}

func (s *UserService) List(ctx context.Context, pagination *database.Pagination) ([]UserResponse, error) {
	users, err := s.userRepo.FindWithPagination(ctx, pagination, nil)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	result := make([]UserResponse, len(users))
	for i, user := range users {
		result[i] = *s.toResponse(&user)
	}

	return result, nil
}

func (s *UserService) Update(ctx context.Context, id uint, req *UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrRecordNotFound, "user not found")
		}
		return errors.Wrap(errors.ErrDatabase, err)
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.Wrap(errors.ErrDatabase, err)
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrRecordNotFound, "user not found")
		}
		return errors.Wrap(errors.ErrDatabase, err)
	}
	return nil
}

func (s *UserService) toResponse(user *model.User) *UserResponse {
	return &UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   user.Status,
	}
}

func hashPassword(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *UserService) Register(ctx context.Context, req *CreateUserRequest) (*auth.TokenPair, error) {
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}
	if exists {
		return nil, errors.New(errors.ErrRecordExists, "username already exists")
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}
	if exists {
		return nil, errors.New(errors.ErrRecordExists, "email already exists")
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashPassword(req.Password),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	return s.jwtAuth.GenerateTokenPair(user.ID, user.Username)
}

func (s *UserService) Login(ctx context.Context, username, password string) (*auth.TokenPair, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrUnauthorized, "invalid username or password")
		}
		return nil, errors.Wrap(errors.ErrDatabase, err)
	}

	if user.Password != hashPassword(password) {
		return nil, errors.New(errors.ErrUnauthorized, "invalid username or password")
	}

	if user.Status != 1 {
		return nil, errors.New(errors.ErrForbidden, "user is disabled")
	}

	return s.jwtAuth.GenerateTokenPair(user.ID, user.Username)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return s.jwtAuth.RefreshAccessToken(refreshToken)
}
