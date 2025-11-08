package http

import (
	"net/http"

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
func (h *AdminUserHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	paginationQuery := util.GetPaginationQuery(r)

	users, pagination, err := h.userUsecase.ListAdmins(r.Context(), paginationQuery)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Gagal mengambil data", err.Error())
		return
	}

	util.SuccessResponseWithPagination(w, "Data admin berhasil diambil", users, &pagination)
}