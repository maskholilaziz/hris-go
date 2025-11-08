package util

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any 		`json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	TotalItems	int64	`json:"total_items"`
	TotalPages	int		`json:"total_pages"`
	CurrentPage	int		`json:"current_page"`
	Limit		int		`json:"limit"`
}

func writeJSON(w http.ResponseWriter, statusCode int, payload JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func SuccessResponse(w http.ResponseWriter, message string, data any) {
	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: message,
		Data: data,
	})
}

func SuccessResponseWithPagination(w http.ResponseWriter, message string, data any, pagination *Pagination) {
	writeJSON(w, http.StatusOK, JSONResponse{
		Success: true,
		Message: message,
		Data: data,
		Pagination: pagination,
	})
}

func ErrorResponse(w http.ResponseWriter, statusCode int, message, err string) {
	writeJSON(w, statusCode, JSONResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func ValidationErrorResponse(w http.ResponseWriter, validationErrors any) {
	writeJSON(w, http.StatusUnprocessableEntity, JSONResponse{
		Success: false,
		Message: "Input tidak valid.",
		Error:   "Validation failed",
		Data:    validationErrors, // Kirim detail error validasinya
	})
}