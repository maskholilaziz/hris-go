package util

import (
	"math"
	"net/http"
	"strconv"
	"strings"
)

type PaginationQuery struct {
	Page		int
	Limit		int
	SortBy		string
	SortDir		string
	Search		string
	Filters		map[string]any
}

func GetPaginationQuery(r *http.Request) PaginationQuery {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	sort := query.Get("sort")
	sortBy := "created_at"
	sortDir := "desc"
	if sort != "" {
		parts := strings.Split(sort, ":")
		sortBy = parts[0]
		if len(parts) > 1 && (parts[1] == "asc" || parts[1] == "desc") {
			sortDir = parts[1]
		}
	}

	search := query.Get("search")

	filters := make(map[string]any)
	for key, values := range query {
		// Kita anggap 'reserved keys' bukan filter
		if key == "page" || key == "limit" || key == "sort" || key == "search" {
			continue
		}
		// Ambil nilai pertama saja
		if len(values) > 0 {
			filters[key] = values[0]
		}
	}

	return PaginationQuery{
		Page:		page,
		Limit:		limit,
		SortBy:		sortBy,
		SortDir:	sortDir,
		Search:		search,
		Filters: 	filters,
	}
}

func (q *PaginationQuery) CalculatePaginationMetadata(totalItems int64) Pagination {
	totalPages := 0
	if totalItems > 0 && q.Limit > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(q.Limit)))
	}

	return Pagination{
		TotalItems: totalItems,
		TotalPages: totalPages,
		CurrentPage: q.Page,
		Limit: q.Limit,
	}
}

func (q *PaginationQuery) GetOffset() int {
	return (q.Page - 1) * q.Limit
}