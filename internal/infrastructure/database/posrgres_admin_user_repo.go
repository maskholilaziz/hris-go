package database

import (
	"context"
	"errors" // Import errors
	"fmt"
	"log" // Import log

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maskholilaziz/hris-go/internal/entity"
	"github.com/maskholilaziz/hris-go/internal/repository"
	"github.com/maskholilaziz/hris-go/pkg/util"

	sq "github.com/Masterminds/squirrel" // 'sq' adalah alias umum
)

// postgresAdminUserRepo adalah implementasi konkret dari AdminUserRepository
type postgresAdminUserRepo struct {
	db  *pgxpool.Pool     // Koneksi pool
	sqb sq.StatementBuilderType // SQL query builder (Squirrel)
}

// NewPostgresAdminUserRepo adalah constructor
func NewPostgresAdminUserRepo(dbPool *pgxpool.Pool) repository.AdminUserRepository {
	return &postgresAdminUserRepo{
		db:  dbPool,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar), // Format $1, $2 untuk Postgres
	}
}

// Create (Contoh menggunakan pgx biasa)
func (r *postgresAdminUserRepo) Create(ctx context.Context, user *entity.AdminUser) error {
	// Pastikan ID sudah di-generate (kita akan generate di usecase)
	if user.ID == uuid.Nil {
		return errors.New("ID user tidak boleh nil")
	}
	
	query := `INSERT INTO admin_users (id, name, email, password, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// FindByEmail (Contoh menggunakan pgx biasa)
func (r *postgresAdminUserRepo) FindByEmail(ctx context.Context, email string) (*entity.AdminUser, error) {
	query := `SELECT id, name, email, password, created_at, updated_at
			  FROM admin_users
			  WHERE email = $1 AND deleted_at IS NULL`
	
	row := r.db.QueryRow(ctx, query, email)
	
	var user entity.AdminUser
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			// 'Best practice': Kembalikan error spesifik jika tidak ditemukan
			return nil, errors.New("admin user not found")
		}
		return nil, err
	}
	
	return &user, nil
}

// FindByID (Sama seperti FindByEmail, tapi pakai ID)
func (r *postgresAdminUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.AdminUser, error) {
	// Implementasi mirip FindByEmail, tapi 'WHERE id = $1'
	// ... (Dipersingkat untuk contoh ini)
	return nil, errors.New("not implemented")
}

// --- Ini adalah bagian untuk 'getAllAdmin' (Filter, Paginate, Sort) ---

// Find (Menggunakan Squirrel)
func (r *postgresAdminUserRepo) Find(ctx context.Context, query util.PaginationQuery) ([]*entity.AdminUser, error) {
	// Mulai membangun query SELECT
	sqlBuilder := r.sqb.Select("id", "name", "email", "created_at", "updated_at").
		From("admin_users").
		Where("deleted_at IS NULL") // Selalu filter soft delete

	// 1. Terapkan Filter (Search)
	if query.Search != "" {
		// Kita cari di kolom 'name' ATAU 'email'
		// 'ILIKE' adalah 'LIKE' yang case-insensitive (khusus Postgres)
		sqlBuilder = sqlBuilder.Where(
			sq.Or{
				sq.ILike{"name": "%" + query.Search + "%"},
				sq.ILike{"email": "%" + query.Search + "%"},
			},
		)
	}
	
	// 2. Terapkan Sorting
	// Kita sudah mem-validasi/default 'SortBy' dan 'SortDir' di helper
	sqlBuilder = sqlBuilder.OrderBy(query.SortBy + " " + query.SortDir)
	
	// 3. Terapkan Pagination
	sqlBuilder = sqlBuilder.Limit(uint64(query.Limit)).
		Offset(uint64(query.GetOffset()))
		
	// Ubah builder menjadi string SQL dan argumennya
	sql, args, err := sqlBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("gagal membangun SQL query: %w", err)
	}

	log.Printf("Executing SQL: %s with args: %v", sql, args)
	
	// Eksekusi query
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan hasilnya
	users := []*entity.AdminUser{}
	for rows.Next() {
		var user entity.AdminUser
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// Count (Menggunakan Squirrel)
func (r *postgresAdminUserRepo) Count(ctx context.Context, query util.PaginationQuery) (int64, error) {
	// Query ini HARUS mencerminkan filter 'Find'
	sqlBuilder := r.sqb.Select("COUNT(*)").
		From("admin_users").
		Where("deleted_at IS NULL")

	// 1. Terapkan Filter (Search) - HARUS SAMA DENGAN 'Find'
	if query.Search != "" {
		sqlBuilder = sqlBuilder.Where(
			sq.Or{
				sq.ILike{"name": "%" + query.Search + "%"},
				sq.ILike{"email": "%" + query.Search + "%"},
			},
		)
	}
	
	sql, args, err := sqlBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("gagal membangun SQL count: %w", err)
	}

	var count int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&count)
	return count, err
}