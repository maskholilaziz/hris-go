package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/usecase"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type CreateTenantRequest struct {
	Name         string `json:"name" validate:"required,min=3,no_consecutive_spaces"`
	CompanyEmail string `json:"company_email" validate:"required,email"`
}

type UpdateTenantRequest struct {
	Name         string `json:"name" validate:"omitempty,min=3,no_consecutive_spaces"`
	CompanyEmail string `json:"company_email" validate:"omitempty,email"`
	Status       string `json:"status" validate:"omitempty,oneof=active inactive suspended setup_pending"`
}

type TenantHandler struct {
	usecase *usecase.TenantUsecase
	validate *validator.Validate
}

func NewTenantHandler(uc *usecase.TenantUsecase, v *validator.Validate) *TenantHandler {
	return &TenantHandler{
		usecase: uc,
		validate: v,
	}
}

type TenantResponse struct {
	ID           uuid.UUID            `json:"id"`
	Name         string               `json:"name"`
	Slug         string               `json:"slug"`
	CompanyEmail string               `json:"company_email"`
	Status       entity.StatusTenant  `json:"status"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	DeletedAt    *time.Time           `json:"deleted_at,omitempty"`
}

type ListTenantsResponse struct {
	Data       []TenantResponse `json:"data"`
	Pagination util.Pagination  `json:"pagination"`
}

func newTenantResponse(t *entity.Tenant) TenantResponse {
	return TenantResponse{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		CompanyEmail: t.CompanyEmail,
		Status:       t.Status,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
		DeletedAt:    t.DeletedAt,
	}
}

func newTenantListResponse(tenants []*entity.Tenant) []TenantResponse {
	responses := make([]TenantResponse, len(tenants))
	for i, t := range tenants {
		responses[i] = newTenantResponse(t)
	}
	return responses
}

func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Input JSON tidak valid", err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.CompanyEmail = strings.TrimSpace(req.CompanyEmail)

	if err := h.validate.Struct(req); err != nil {
		errors := util.FormatValidationErrors(err.(validator.ValidationErrors))
		util.ErrorResponse(w, http.StatusUnprocessableEntity, "Input tidak valid", fmt.Sprintf("%v", errors))
		return
	}

	input := usecase.CreateTenantInput{
		Name:         req.Name,
		CompanyEmail: req.CompanyEmail,
	}

	tenant, err := h.usecase.CreateTenant(r.Context(), input)
	if err != nil {
		if err.Error() == "nama tenant sudah terdaftar" {
			util.ErrorResponse(w, http.StatusConflict, "Registrasi gagal", err.Error())
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, "Gagal membuat tenant", err.Error())
		return
	}

	util.SuccessResponse(w, "Tenant berhasil dibuat", tenant)
}

func (h *TenantHandler) List(w http.ResponseWriter, r *http.Request) {
	paginationQuery := util.GetPaginationQuery(r)

	tenants, pagination, err := h.usecase.ListTenants(r.Context(), paginationQuery)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Gagal mengambil data tenant", err.Error())
		return
	}

	response := ListTenantsResponse{
		Data:       newTenantListResponse(tenants),
		Pagination: pagination,
	}

	util.SuccessResponse(w, "Data tenant berhasil diambil", response)
}

func (h *TenantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "ID tenant tidak valid", err.Error())
		return
	}

	tenant, err := h.usecase.GetTenantByID(r.Context(), id)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Tenant tidak ditemukan", err.Error())
		return
	}

	util.SuccessResponse(w, "Tenant berhasil diambil", newTenantResponse(tenant))
}

func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "ID tenant tidak valid", err.Error())
		return
	}

	var req UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Input JSON tidak valid", err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.CompanyEmail = strings.TrimSpace(req.CompanyEmail)
	req.Status = strings.TrimSpace(req.Status)

	if err := h.validate.Struct(req); err != nil {
		errors := util.FormatValidationErrors(err.(validator.ValidationErrors))
		util.ErrorResponse(w, http.StatusUnprocessableEntity, "Input tidak valid", fmt.Sprintf("%v", errors))
		return
	}

	input := usecase.UpdateTenantInput{
		Name:         req.Name,
		CompanyEmail: req.CompanyEmail,
		Status:       req.Status,
	}

	tenant, err := h.usecase.UpdateTenant(r.Context(), id, input)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Gagal memperbarui tenant", err.Error())
		return
	}

	util.SuccessResponse(w, "Tenant berhasil diupdate", newTenantResponse(tenant))
}

func (h *TenantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "ID tenant tidak valid", err.Error())
		return
	}

	err = h.usecase.DeleteTenant(r.Context(), id)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Gagal menghapus tenant", err.Error())
		return
	}
	
	util.SuccessResponse(w, "Tenant berhasil dihapus", nil)
}