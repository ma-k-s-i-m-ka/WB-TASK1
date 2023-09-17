package main

import (
	"WBL0/app/internal/server"
	"WBL0/app/pkg/config"
	"WBL0/app/pkg/logger"
	"WBL0/app/pkg/nats"
	postgres "WBL0/app/pkg/storage"

	"context"
	"errors"
	"flag"
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

	natsConn, err := nats.ConnectNATS(*cfg)
	if err != nil {
		log.Fatal("cannot connect to NATS:", err)
	}
	log.Info("connected to NATS")

	router := httprouter.New()
	log.Info("initialized httprouter")

	srv := server.NewServer(cfg, router, &log)
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
