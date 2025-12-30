package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"toDoList/internal"
	"toDoList/internal/repository/db"
	"toDoList/internal/repository/inmemory"
	"toDoList/internal/server"
	auth "toDoList/internal/server/auth/user_auth"
	"toDoList/internal/server/workers"
	"toDoList/pkg/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		defer signal.Stop(c)
		<-c
		cancel()
	}()

	cfg := internal.ReadConfig()

	log := logger.Init(cfg.Debug)
	log.Debug().Any("config", cfg).Send()

	log.Info().Msg("Server starting...")

	var database server.Storage

	// конфигураця и создание хранилища
	postgresDB, err := db.NewStorage(cfg.DNS)
	if err != nil {
		log.Err(err).Msg("Postgres недоступен (%v), используем in-memory storage")
		inmemoryDB := inmemory.NewInMemoryStorage()
		database = inmemoryDB
	} else {
		database = postgresDB
		// запуск миграции
		if err = db.Migrations(cfg.DNS, cfg.MigratePath); err != nil {
			cancel()
			log.Error().Err(err).Msg("failed to migrate db")
			return
		}
	}
	taskDeleter := workers.NewTaskBatchDeleter(ctx, database, cfg.TaskCapacity, log)

	signer := auth.HS256Signer{
		Secret:     []byte("ultraSecretKey123"),
		Issuer:     "todolistService",
		Audience:   "todolistClient",
		AccessTTL:  internal.MinFifteen,
		RefreshTTL: internal.OneWeek,
	}

	srv := server.NewServer(cfg, database, signer, taskDeleter)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// конфигурация и запуск веб-сервера
		if err = srv.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		taskDeleter.Start()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), internal.SecTen)
		defer shutdownCancel()
		err = srv.ShutDown(shutdownCtx)
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown server")
		}
		err = taskDeleter.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to delete marked tasks")
		}
		err = postgresDB.Close(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown database")
		}
		log.Info().Msg("Server stopped")
	}()

	wg.Wait()
}
