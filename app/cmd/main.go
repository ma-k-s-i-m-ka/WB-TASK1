package main

import (
	"WBL0/app/internal/cache"
	"WBL0/app/internal/delivery"
	"WBL0/app/internal/item"
	"WBL0/app/internal/order"
	"WBL0/app/internal/payment"
	"WBL0/app/internal/server"
	"WBL0/app/pkg/config"
	"WBL0/app/pkg/logger"
	"WBL0/app/pkg/nats"
	postgres "WBL0/app/pkg/storage"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := logger.GetLogger()
	log.Info("logger initialized")

	configPath := flag.String("config-path", "config.yml", "path for application configuration file")
	cfg := config.GetConfig(*configPath, ".env")
	log.Info("loaded config file")

	dbConn, err := postgres.ConnectDB(*cfg)
	if err != nil {
		log.Error("cannot connect to database", err)
	}
	log.Info("connected to database")

	allCache := cache.NewCache()
	if err = loadAllCache(log, dbConn, allCache); err != nil {
		log.Error("Failed to preload caches:", err)
	}
	log.Info("records from the database are added to the cache")

	fmt.Println("Cache after loading data:")
	for key, value := range allCache.Payments {
		fmt.Printf("Key: %d, Value: %+v\n", key, value)
	}

	natsConn, err := nats.ConnectNATS(*cfg)
	if err != nil {
		log.Fatal("cannot connect to NATS:", err)
	}
	log.Info("connected to NATS")

	router := httprouter.New()
	log.Info("initialized httprouter")

	srv := server.NewServer(cfg, router, &log, allCache)
	log.Info("starting the server")

	quit := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM}
	signal.Notify(quit, signals...)

	go func() {
		if err = srv.Run(dbConn, natsConn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("cannot run the server", err)
		}
	}()
	log.Info("server has been started ", slog.String("host", cfg.HTTP.Host), slog.String("port", cfg.HTTP.Port))

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		dbCloseCtx, dbCloseCancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.PostgreSQL.ShutdownTimeout)*time.Second,
		)
		defer dbCloseCancel()
		err = dbConn.Close(dbCloseCtx)
		if err != nil {
			log.Error("failed to close database connection:", err)
		}
		log.Info("closed database connection")

		natsConn.Close()
		if err = natsConn.LastError(); err != nil {
			log.Error("failed to close NATS connection:", err)
		}
		log.Info("closed NATS connection")
		cancel()
	}()

	if err = srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed:", err)
	}
	log.Info("server has been shutted down")
}

func loadAllCache(log logger.Logger, dbConn *pgx.Conn, cache *cache.Cache) error {
	if err := delivery.CacheForDelivery(dbConn, cache); err != nil {
		log.Error("Failed to load delivery data into cache:", err)
		return err
	}
	if err := order.CacheForOrder(dbConn, cache); err != nil {
		log.Error("Failed to load order data into cache:", err)
		return err
	}

	if err := item.CacheForItem(dbConn, cache); err != nil {
		log.Error("Failed to load item data into cache:", err)
		return err
	}
	if err := payment.CacheForPayment(dbConn, cache); err != nil {
		log.Error("Failed to load payment data into cache:", err)
		return err
	}
	return nil
}
