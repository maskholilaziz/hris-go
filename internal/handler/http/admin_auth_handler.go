package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/maskholilaziz/hris-go/internal/usecase"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type AdminAuthHandler struct {
	authUsecase *usecase.AdminAuthUsecase
	validate	*validator.Validate
}

func NewAdminAuthHandler(uc *usecase.AdminAuthUsecase, v *validator.Validate) *AdminAuthHandler {
	return &AdminAuthHandler{
		authUsecase: uc,
		validate: v,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"` // Hanya perlu 'required' saat login
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=10,no_consecutive_spaces"`
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			util.ErrorResponse(w, http.StatusBadRequest, "Request body tidak boleh kosong", err.Error())
			return
		}
		util.ErrorResponse(w, http.StatusBadRequest, "Input JSON tidak valid", err.Error())
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			util.ErrorResponse(w, http.StatusInternalServerError, "Gagal memvalidasi input", err.Error())
			return
		}
		
		errors := util.FormatValidationErrors(validationErrors)
		util.ErrorResponse(w, http.StatusUnprocessableEntity, "Input tidak valid", fmt.Sprintf("%v", errors))
		return
	}

	token, err := h.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, "Login gagal", err.Error())
		return
	}

	util.SuccessResponse(w, "Login berhasil", map[string]string{
		"token": token,
	})
}

func (h *AdminAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			util.ErrorResponse(w, http.StatusBadRequest, "Request body tidak boleh kosong", err.Error())
			return
		}
		util.ErrorResponse(w, http.StatusBadRequest, "Input JSON tidak valid", err.Error())
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)
	req.Password = strings.TrimSpace(req.Password)

	err = h.validate.Struct(req)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			util.ErrorResponse(w, http.StatusInternalServerError, "Gagal memvalidasi input", err.Error())
			return
		}

		// --- PERBAIKAN POIN 3 ---
		// Panggil helper global dari pkg/util
		errors := util.FormatValidationErrors(validationErrors)
		util.ErrorResponse(w, http.StatusUnprocessableEntity, "Input tidak valid", fmt.Sprintf("%v", errors))
		return
	}

	user, err := h.authUsecase.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Registrasi gagal", err.Error())
		return
	}

	util.SuccessResponse(w, "Registrasi berhasil", user)
}