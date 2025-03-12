package service

import (
	"context"
	"time"

	"github.com/tieubaoca/chatbot-be/repository"
	"github.com/tieubaoca/chatbot-be/types"
)

type UserService interface {
	CreateUser(ctx context.Context, user *types.User) error
	BatchCreateUser(ctx context.Context, users []*types.User) error
	GetUser(ctx context.Context, id string) (*types.User, error)
	GetUserByWorkspace(ctx context.Context, workspace string) ([]*types.User, error)
	UpdateUser(ctx context.Context, id string, user *types.User) error
	DeleteUser(ctx context.Context, id string) error
	PaginateUser(ctx context.Context, page int64, limit int64) ([]*types.User, int64, error)
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
}

type userService struct {
	repo repository.UserRepo
}

func NewUserService(repo repository.UserRepo) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *types.User) error {
	user.CreateAt = time.Now().Unix()
	user.UpdateAt = time.Now().Unix()

	return s.repo.CreateUser(ctx, user)
}

func (s *userService) BatchCreateUser(ctx context.Context, users []*types.User) error {
	for _, user := range users {
		user.CreateAt = time.Now().Unix()
		user.UpdateAt = time.Now().Unix()
	}
	return s.repo.BatchCreateUser(ctx, users)
}

func (s *userService) GetUser(ctx context.Context, id string) (*types.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *userService) GetUserByWorkspace(ctx context.Context, workspace string) ([]*types.User, error) {
	return s.repo.GetUserByWorkspace(ctx, workspace)
}

func (s *userService) UpdateUser(ctx context.Context, id string, user *types.User) error {
	dbUser, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return err
	}
	user.UpdateAt = time.Now().Unix()
	if user.Username != "" {
		dbUser.Username = user.Username
	}
	if user.Password != "" {
		dbUser.Password = user.Password
	}
	if user.FullName != "" {
		dbUser.FullName = user.FullName
	}
	if user.Workspace != "" {
		dbUser.Workspace = user.Workspace
	}
	if user.WorkspaceRole != "" {
		dbUser.WorkspaceRole = user.WorkspaceRole
	}
	if user.ManagementLevel != 0 {
		dbUser.ManagementLevel = user.ManagementLevel
	}

	return s.repo.UpdateUser(ctx, id, dbUser)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *userService) PaginateUser(ctx context.Context, page int64, limit int64) ([]*types.User, int64, error) {
	return s.repo.PaginateUser(ctx, page, limit)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}
