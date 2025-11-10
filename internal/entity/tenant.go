package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type StatusTenant string

const (
	StatusTenantSetupPending StatusTenant = "setup_pending"
	StatusTenantActive       StatusTenant = "active"
	StatusTenantInactive     StatusTenant = "inactive"
	StatusTenantSuspended    StatusTenant = "suspended"
)

type Tenant struct {
	ID           uuid.UUID
	Name         string
	Slug         string
	CompanyEmail string
	Status       StatusTenant
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func NewTenant(name, companyEmail string) *Tenant {
	tenantSlug := slug.Make(name)
	
	return &Tenant{
		ID:           uuid.New(),
		Name:         name,
		Slug:         tenantSlug,
		CompanyEmail: companyEmail,
		Status:       StatusTenantSetupPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (t *Tenant) SetSlug(name string) {
	t.Slug = slug.Make(name)
}