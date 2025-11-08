package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/infrastructure/security"
	"github.com/maskholilaziz/hris-go/internal/repository"
)

type AdminAuthUsecase struct {
	adminRepo repository.AdminUserRepository
	jwtService *security.JWTService
}

func NewAdminAuthUsecase(
	adminRepo repository.AdminUserRepository,
	jwtService *security.JWTService,
) *AdminAuthUsecase {
	return &AdminAuthUsecase{
		adminRepo: adminRepo,
		jwtService: jwtService,
	}
}

func (uc *AdminAuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", errors.New("email atau password salah")
	}

	if !user.CheckPassword(password) {
		return "", errors.New("email atau password salah")
	}

	token, err := uc.jwtService.GenerateSuperadminToken(user.ID)
	if err != nil {
		return "", errors.New("gagal membuat token")
	}

	return token, nil
}

func (uc *AdminAuthUsecase) Register(ctx context.Context, name, email, password string) (*entity.AdminUser, error) {
	if _, err := uc.adminRepo.FindByEmail(ctx, email); err == nil {
		return nil, errors.New("email sudah terdaftar")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, errors.New("gagal membuat UUID")
	}

	user := &entity.AdminUser{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.HashPassword(password); err != nil {
		return nil, errors.New("gagal hash password")
	}

	if err := uc.adminRepo.Create(ctx, user); err != nil {
		return nil, errors.New("gagal menyimpan user")
	}

	user.Password = ""
	return user, nil
}