package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)
	Find(ctx context.Context, query util.PaginationQuery) ([]*entity.Tenant, error)
	Count(ctx context.Context, query util.PaginationQuery) (int64, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}