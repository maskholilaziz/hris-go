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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maskholilaziz/hris-go/internal/config"
	"github.com/maskholilaziz/hris-go/internal/infrastructure/database"
)

func main() {
	log.Println("Menjalankan server...")

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Tidak bisa memuat konfigurasi: %v", err)
	}

	dbPool := database.NewDBConnection(cfg.DatabaseURL)

	defer dbPool.Close()

	r := chi.NewRouter()
	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func (w http.ResponseWriter, r *http.Request)  {
		response := map[string]string{
			"status": "OK",
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(response)
	})

	r.Get("/ready", readinessCheckHandler(dbPool))

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func()  {
		log.Printf("Server berjalan di http://localhost:%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Tidak bisa menjalankan server: %v", err)
		}
	} ()

	<- stop

	log.Println("Server menerima sinyal untuk berhenti...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown gagal: %v", err)
	}

	log.Println("Server berhenti dengan sukses.")
}

func readinessCheckHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := dbPool.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"db":     "unreachable",
			})

			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
			"db":     "reachable",
		})
	}
}