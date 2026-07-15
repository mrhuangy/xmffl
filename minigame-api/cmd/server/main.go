package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fpxxl/minigame-api/internal/config"
	"fpxxl/minigame-api/internal/database"
	"fpxxl/minigame-api/internal/handler"
	"fpxxl/minigame-api/internal/repository"
	"fpxxl/minigame-api/internal/router"
	"fpxxl/minigame-api/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := database.Wait(ctx, db); err != nil {
		log.Fatalf("wait database: %v", err)
	}

	repo := repository.NewMySQLRepository(db)
	playerService := service.NewPlayerService(repo)
	gameService := service.NewGameService(repo)
	shopService := service.NewShopService(repo)
	apiHandler := handler.New(repo, playerService, gameService, shopService)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router.New(cfg, repo, apiHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("minigame api listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
