package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"paintwar/server/internal/config"
	"paintwar/server/internal/db"
	"paintwar/server/internal/hub"
	"paintwar/server/internal/persist"
	"paintwar/server/internal/ws"
)

func main() {
	cfg := config.Load()

	rootCtx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()

	// Persistence is optional: without DATABASE_URL the game still runs.
	var persister hub.Persister
	var writer *persist.Writer
	if cfg.DatabaseURL != "" {
		pool, err := db.NewPool(rootCtx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		defer pool.Close()
		store := db.NewStore(pool)
		if err := store.Migrate(rootCtx); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		writer = persist.New(store)
		persister = writer
		log.Println("persistence enabled")
	} else {
		log.Println("DATABASE_URL not set; persistence disabled")
	}

	h := hub.New(int64(cfg.MatchDurationMs), persister)
	if writer != nil {
		writer.Start(rootCtx, h.UpdateLeaderboard)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ws", ws.NewHandler(h))

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	cancelRoot()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
