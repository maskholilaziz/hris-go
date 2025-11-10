package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/usecase"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type AdminUserHandler struct {
	userUsecase *usecase.AdminUserUsecase
}

func NewAdminUserHandler(uc *usecase.AdminUserUsecase) *AdminUserHandler {
	return &AdminUserHandler{
		userUsecase: uc,
	}
}

type AdminUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListAdminUsersResponse struct {
	Data       []AdminUserResponse `json:"data"`
	Pagination util.Pagination     `json:"pagination"`
}

func newAdminUserResponse(user *entity.AdminUser) AdminUserResponse {
	return AdminUserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func newListAdminUsersResponse(users []*entity.AdminUser) []AdminUserResponse {
	responses := make([]AdminUserResponse, len(users))
	for i, user := range users {
		responses[i] = newAdminUserResponse(user)
	}
	return responses
}

func (h *AdminUserHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	paginationQuery := util.GetPaginationQuery(r)

	users, pagination, err := h.userUsecase.ListAdmins(r.Context(), paginationQuery)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Gagal mengambil data", err.Error())
		return
	}

	response := ListAdminUsersResponse{
		Data:       newListAdminUsersResponse(users),
		Pagination: pagination,
	}
	util.SuccessResponse(w, "Data admin berhasil diambil", response)
}