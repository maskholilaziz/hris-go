package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBConnection membuat dan mengembalikan sebuah connection pool ke database PostgreSQL.
// Kita menggunakan *pgxpool.Pool, yang merupakan cara "best practice" dari pgx
// untuk mengelola koneksi. Ini thread-safe dan menangani koneksi yang sibuk/idle
// secara otomatis.
func NewDBConnection(databaseURL string) *pgxpool.Pool {
	// Konfigurasi parsing. Ini adalah langkah pertama untuk memberitahu pgx
	// cara terhubung ke database.
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Tidak dapat parse konfigurasi database: %v", err)
	}

	// (Opsional) Ini adalah beberapa pengaturan "best practice" untuk connection pool
	// agar mumpuni untuk skala besar.
	config.MaxConns = 10           // Jumlah maksimum koneksi
	config.MinConns = 2            // Jumlah minimum koneksi
	config.MaxConnIdleTime = time.Minute * 30 // Waktu idle sebelum koneksi ditutup
	config.MaxConnLifetime = time.Hour * 1    // Waktu hidup maksimum koneksi
	config.HealthCheckPeriod = time.Minute * 1 // Seberapa sering melakukan health check

	// Membuat context untuk proses koneksi.
	// Kita beri timeout 5 detik. Jika DB tidak merespon dalam 5 detik,
	// aplikasi akan gagal start (fail-fast).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mencoba membuat connection pool.
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Tidak dapat membuat connection pool: %v", err)
	}

	// Melakukan 'Ping' ke database untuk memastikan koneksi benar-benar
	// berhasil dan database siap menerima query.
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Tidak dapat ping database: %v", err)
	}

	log.Println("Koneksi database berhasil dibuat.")
	return pool
}