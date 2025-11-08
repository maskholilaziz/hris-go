package usecase

import (
	"context"

	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/repository"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type AdminUserUsecase struct {
	adminRepo repository.AdminUserRepository
}

func NewAdminUserUsecase(adminRepo repository.AdminUserRepository) *AdminUserUsecase {
	return &AdminUserUsecase{
		adminRepo: adminRepo,
	}
}

func (uc *AdminUserUsecase) ListAdmins(ctx context.Context, query util.PaginationQuery) ([]*entity.AdminUser, util.Pagination, error) {
	users, err := uc.adminRepo.Find(ctx, query)
	if err != nil {
		return nil, util.Pagination{}, err
	}

	totalItems, err := uc.adminRepo.Count(ctx, query)
	if err != nil {
		return nil, util.Pagination{}, err
	}

	pagination := query.CalculatePaginationMetadata(totalItems)

	return users, pagination, nil
}