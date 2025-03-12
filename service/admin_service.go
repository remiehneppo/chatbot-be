package service

import (
	"context"

	"github.com/tieubaoca/chatbot-be/repository"
	"github.com/tieubaoca/chatbot-be/types"
)

type AdminService interface {
	GetAdmin(ctx context.Context, id string) (*types.Admin, error)
	GetAdminByUsername(ctx context.Context, username string) (*types.Admin, error)
	CreateAdmin(ctx context.Context, admin *types.Admin) error
	UpdateAdmin(ctx context.Context, id string, admin *types.Admin) error
	DeleteAdmin(ctx context.Context, id string) error
}

type adminService struct {
	repo repository.AdminRepo
}

func NewAdminService(repo repository.AdminRepo) AdminService {
	return &adminService{
		repo: repo,
	}
}

func (s *adminService) GetAdmin(ctx context.Context, id string) (*types.Admin, error) {
	return s.repo.GetAdmin(ctx, id)
}

func (s *adminService) GetAdminByUsername(ctx context.Context, username string) (*types.Admin, error) {
	return s.repo.GetAdminByUsername(ctx, username)
}

func (s *adminService) CreateAdmin(ctx context.Context, admin *types.Admin) error {
	return s.repo.CreateAdmin(ctx, admin)
}

func (s *adminService) UpdateAdmin(ctx context.Context, id string, admin *types.Admin) error {
	dbAdmin, err := s.repo.GetAdmin(ctx, id)
	if err != nil {
		return err
	}
	if admin.Username != "" {
		dbAdmin.Username = admin.Username
	}
	if admin.Password != "" {
		dbAdmin.Password = admin.Password
	}
	if admin.Role != "" {
		dbAdmin.Role = admin.Role
	}

	return s.repo.UpdateAdmin(ctx, id, dbAdmin)
}

func (s *adminService) DeleteAdmin(ctx context.Context, id string) error {
	return s.repo.DeleteAdmin(ctx, id)
}
