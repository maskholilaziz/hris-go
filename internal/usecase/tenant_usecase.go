package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/repository"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type CreateTenantInput struct {
	Name         string `json:"name"`
	CompanyEmail string `json:"company_email"`
}

type UpdateTenantInput struct {
	Name         string `json:"name"`
	CompanyEmail string `json:"company_email"`
	Status       string `json:"status"`
}

type TenantUsecase struct {
	tenantRepo repository.TenantRepository
}

func NewTenantUsecase(tenantRepo repository.TenantRepository) *TenantUsecase {
	return &TenantUsecase{
		tenantRepo: tenantRepo,
	}
}

func (uc *TenantUsecase) CreateTenant(ctx context.Context, input CreateTenantInput) (*entity.Tenant, error) {
	tenant := entity.NewTenant(input.Name, input.CompanyEmail)
	
	if existing, _ := uc.tenantRepo.FindBySlug(ctx, tenant.Slug); existing != nil {
		return nil, errors.New("nama tenant sudah terdaftar")
	}

	if err := uc.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("gagal menyimpan tenant: %w", err)
	}

	return tenant, nil
}

func (uc *TenantUsecase) ListTenants(ctx context.Context, query util.PaginationQuery) ([]*entity.Tenant, util.Pagination, error) {
	tenant, err := uc.tenantRepo.Find(ctx, query)
	if err != nil {
		return nil, util.Pagination{}, err
	}

	totalItems, err := uc.tenantRepo.Count(ctx, query)
	if err != nil {
		return nil, util.Pagination{}, err
	}

	pagination := query.CalculatePaginationMetadata(totalItems)
	
	return tenant, pagination, nil
}

func (uc *TenantUsecase) GetTenantByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	return uc.tenantRepo.FindByID(ctx, id)
}

func (uc *TenantUsecase) UpdateTenant(ctx context.Context, id uuid.UUID, input UpdateTenantInput) (*entity.Tenant, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" && input.Name != tenant.Name {
		tenant.Name = input.Name

		newSlug := slug.Make(input.Name)
		if newSlug != tenant.Slug {
			if existing, _ := uc.tenantRepo.FindBySlug(ctx, newSlug); existing != nil {
				return nil, errors.New("nama tenant sudah ada (slug conflict)")
			}
		}
	}

	if input.CompanyEmail != "" {
		tenant.CompanyEmail = input.CompanyEmail
	}

	if input.Status != "" {
		s := entity.StatusTenant(input.Status)
		if s != entity.StatusTenantActive && s != entity.StatusTenantInactive && s != entity.StatusTenantSetupPending && s != entity.StatusTenantSuspended {
			return nil, errors.New("status tenant tidak valid")
		}
		tenant.Status = s
	}

	if err := uc.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil 
}

func (uc *TenantUsecase) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	if _, err := uc.tenantRepo.FindByID(ctx, id); err != nil {
		return err
	}
	
	return uc.tenantRepo.Delete(ctx, id)
}