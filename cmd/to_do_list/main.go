package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"toDoList/internal"
	"toDoList/internal/repository/db"
	"toDoList/internal/repository/inmemory"
	"toDoList/internal/server"

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

	// конфигураця приложения
	fmt.Println("To-do-list Api is starting")
	cfg := internal.ReadConfig()

	var database server.Storage

	// конфигураця и создание хранилища
	postgresDB, err := db.NewStorage(cfg.DNS)
	if err != nil {
		log.Printf("Postgres недоступен (%v), используем in-memory storage", err)
		inmemoryDB := inmemory.NewInMemoryStorage()
		database = inmemoryDB
	} else {
		database = postgresDB
		// запуск миграции
		if err = db.Migrations(cfg.DNS, cfg.MigratePath); err != nil {
			log.Fatal(err)
		}
	}

	srv := server.NewServer(cfg, database)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// конфигурация и запуск веб-сервера
		if err = srv.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err = srv.ShutDown(ctx)
		if err != nil {
			log.Fatal(err)
		}
		err = postgresDB.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("To-do-list Api is shutting down")
	}()
	wg.Wait()
}
