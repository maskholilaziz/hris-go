package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/repository"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

type postgresTenantRepo struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func NewPostgresTenantRepo(dbPool *pgxpool.Pool) repository.TenantRepository {
	return &postgresTenantRepo{
		db:  dbPool,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *postgresTenantRepo) Create(ctx context.Context, tenant *entity.Tenant) error {
	query := `INSERT INTO tenants (id, name, slug, company_email, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.CompanyEmail,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	return err
}

func (r *postgresTenantRepo) scanTenant(row pgx.Row) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := row.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.CompanyEmail,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *postgresTenantRepo) FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	query := `SELECT id, name, slug, company_email, status, created_at, updated_at
			  FROM tenants
			  WHERE slug = $1 AND deleted_at IS NULL`
	
	row := r.db.QueryRow(ctx, query, slug)
	return r.scanTenant(row)
}

func (r *postgresTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	query := `SELECT id, name, slug, company_email, status, created_at, updated_at
			  FROM tenants
			  WHERE id = $1 AND deleted_at IS NULL`
	
	row := r.db.QueryRow(ctx, query, id)
	return r.scanTenant(row)
}

func (r *postgresTenantRepo) buildFindQuery(query util.PaginationQuery, isCount bool) (string, []interface{}, error) {
	var sb sq.SelectBuilder
	if isCount {
		sb = r.sqb.Select("COUNT(*)").From("tenants")
	} else {
		sb = r.sqb.Select("id", "name", "slug", "company_email", "status", "created_at", "updated_at").
			From("tenants")
	}

	sb = sb.Where("deleted_at IS NULL")

	if query.Search != "" {
		sb = sb.Where(
			sq.Or{
				sq.ILike{"name": "%" + query.Search + "%"},
				sq.ILike{"company_email": "%" + query.Search + "%"},
				sq.ILike{"slug": "%" + query.Search + "%"},
			},
		)
	}

	if query.Filters != nil {
		if status, ok := query.Filters["status"].(string); ok && status != "" {
			sb = sb.Where(sq.Eq{"status": status})
		}
	}

	if !isCount {
		// Terapkan Sorting
		sb = sb.OrderBy(query.SortBy + " " + query.SortDir)
		
		// Terapkan Pagination
		sb = sb.Limit(uint64(query.Limit)).
			Offset(uint64(query.GetOffset()))
	}
	
	return sb.ToSql()
}

func (r *postgresTenantRepo) Find(ctx context.Context, query util.PaginationQuery) ([]*entity.Tenant, error) {
	sql, args, err := r.buildFindQuery(query, false)
	if err != nil {
		return nil, fmt.Errorf("gagal membangun SQL query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := []*entity.Tenant{}
	for rows.Next() {
		tenant, err := r.scanTenant(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

func (r *postgresTenantRepo) Count(ctx context.Context, query util.PaginationQuery) (int64, error) {
	sql, args, err := r.buildFindQuery(query, true)
	if err != nil {
		return 0, fmt.Errorf("gagal membangun SQL count: %w", err)
	}

	var count int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&count)
	return count, err
}

func (r *postgresTenantRepo) Update(ctx context.Context, tenant *entity.Tenant) error {
	query := `UPDATE tenants
			  SET name = $1, slug = $2, company_email = $3, status = $4, updated_at = $5
			  WHERE id = $6 AND deleted_at IS NULL`
	
	tenant.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(ctx, query,
		tenant.Name,
		tenant.Slug,
		tenant.CompanyEmail,
		tenant.Status,
		tenant.UpdatedAt,
		tenant.ID,
	)
	return err
}

func (r *postgresTenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenants SET deleted_at = $1, updated_at = $1 WHERE id = $2`
	now := time.Now()
	
	_, err := r.db.Exec(ctx, query, now, id)
	return err
}