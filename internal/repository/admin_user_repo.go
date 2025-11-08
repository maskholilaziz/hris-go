package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type AdminUserRepository interface {
	Create(ctx context.Context, user *entity.AdminUser) error
	FindByEmail(ctx context.Context, email string) (*entity.AdminUser, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.AdminUser, error)
	Find(ctx context.Context, query util.PaginationQuery) ([]*entity.AdminUser, error)
	Count(ctx context.Context, query util.PaginationQuery) (int64, error)
}