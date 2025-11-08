package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Import library yang kita butuhkan
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	// Import package internal kita

	"github.com/maskholilaziz/hris-go/internal/config"
	inhttp "github.com/maskholilaziz/hris-go/internal/handler/http"
	"github.com/maskholilaziz/hris-go/internal/infrastructure/database"
	"github.com/maskholilaziz/hris-go/internal/infrastructure/security"
	"github.com/maskholilaziz/hris-go/internal/usecase"
	"github.com/maskholilaziz/hris-go/pkg/util"
)

func main() {
	// ------------------------------------------------------------------------
	// Inisialisasi
	// ------------------------------------------------------------------------
	log.Println("Menjalankan server...")

	// 1. Muat Konfigurasi
	// Kita memuat dari "." (root direktori)
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Tidak bisa memuat konfigurasi: %v", err)
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET harus di-set di file .env")
	}

	// 2. Buat Koneksi Database
	// Kita teruskan connection string dari config yang sudah dimuat
	dbPool := database.NewDBConnection(cfg.DatabaseURL)
	// Kita 'defer' Close() untuk memastikan pool ditutup dengan baik saat
	// aplikasi (fungsi main) selesai.
	defer dbPool.Close()

	// 3. Inisialisasi Router (Chi)
	r := chi.NewRouter()
	
	// ------------------------------------------------------------------------
	// Middleware
	// ------------------------------------------------------------------------

	// Pasang middleware "best practice" dari Chi.
	// Logger: Mencatat log setiap request
	// Recoverer: Menangkap panic agar server tidak mati
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	validate := util.NewValidator()

	jwtService := security.NewJWTService(cfg.JWTSecret)

	adminUserRepo := database.NewPostgresAdminUserRepo(dbPool)

	adminAuthUsecase := usecase.NewAdminAuthUsecase(adminUserRepo, jwtService)
	adminUserUsecase := usecase.NewAdminUserUsecase(adminUserRepo)

	adminAuthHandler := inhttp.NewAdminAuthHandler(adminAuthUsecase, validate)
	adminUserHandler := inhttp.NewAdminUserHandler(adminUserUsecase)

	// ------------------------------------------------------------------------
	// Routes / Endpoints
	// ------------------------------------------------------------------------

	// Kita akan buat dua endpoint health check. Ini adalah "best practice":
	// 1. /health (Liveness): Cek apakah aplikasi "hidup" (server web berjalan)
	// 2. /ready (Readiness): Cek apakah aplikasi "siap" bekerja (DB terhubung)
	r.Get("/health", healthCheckHandler)

	// Untuk endpoint ini, kita perlu 'dbPool' untuk ping ke DB.
	// Kita gunakan handler function biasa untuk men-demonstrasikan
	// cara Chi menangani handler.
	r.Get("/ready", readinessCheckHandler(dbPool))

	r.Route("/superadmin", func(r chi.Router) {
		r.Post("/login", adminAuthHandler.Login)
		r.Post("/register", adminAuthHandler.Register)

		r.Group(func(r chi.Router) {
			r.Use(jwtService.SuperadminAuthMiddleware)
			r.Get("/users", adminUserHandler.ListAdmins)
		})
	})

	// ------------------------------------------------------------------------
	// Menjalankan Server
	// ------------------------------------------------------------------------

	// Konfigurasi server HTTP
	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	// Channel untuk "mendengarkan" OS Signal (e.g., Ctrl+C)
	// Ini adalah bagian dari 'graceful shutdown'
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Jalankan server di goroutine terpisah
	go func()  {
		log.Printf("Server berjalan di http://localhost:%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Tidak bisa menjalankan server: %v", err)
		}
	} ()

	// Tunggu sinyal 'stop' (graceful shutdown)
	<- stop

	log.Println("Server menerima sinyal untuk berhenti...")

	// Beri waktu 5 detik untuk request yang sedang berjalan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server dengan 'graceful'
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown gagal: %v", err)
	}

	log.Println("Server berhenti dengan sukses.")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	util.SuccessResponse(w, "Server is alive", nil)
}

// readinessCheckHandler adalah handler yang menerima 'dbPool'.
// Ini adalah "pattern" yang bagus: sebuah fungsi yang *mengembalikan*
// http.HandlerFunc.
func readinessCheckHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Coba ping ke DB dengan timeout 3 detik
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := dbPool.Ping(ctx); err != nil {
			// Jika gagal, kirim status 503 Service Unavailable
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"db":     "unreachable",
			})

			return
		}

		// Jika berhasil, kirim status 200 OK
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
			"db":     "reachable",
		})
	}
}