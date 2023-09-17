package server

import (
	"WBL0/app/internal/cache"
	"WBL0/app/internal/delivery"
	"WBL0/app/internal/item"
	"WBL0/app/internal/model"
	"WBL0/app/internal/order"
	"WBL0/app/internal/payment"
	"WBL0/app/pkg/config"
	"WBL0/app/pkg/logger"
	nats2 "WBL0/app/pkg/nats"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/julienschmidt/httprouter"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
	"time"
)

type Server struct {
	srv     *http.Server
	log     *logger.Logger
	cfg     *config.Config
	handler *httprouter.Router
	cache   *cache.Cache
}

func NewServer(cfg *config.Config, handler *httprouter.Router, log *logger.Logger) *Server {
	cache := &cache.Cache{
		Deliveries: make(map[int64]*model.Delivery),
		Payments:   make(map[int64]*model.Payment),
		Items:      make(map[int64]*model.Item),
		Orders:     make(map[string]*model.Order),
	}
	return &Server{
		srv: &http.Server{
			Handler:      handler,
			WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
			ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
			Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		},
		log:     log,
		cfg:     cfg,
		handler: handler,
		cache:   cache,
	}
}

func (s *Server) Run(dbConn *pgx.Conn, natsConn *nats.Conn) error {

	reqTimeout := s.cfg.PostgreSQL.RequestTimeout

	orderStorage := order.NewStorage(dbConn, reqTimeout, s.cache)
	orderService := order.NewService(orderStorage, *s.log)
	orderHandler := order.NewHandler(*s.cfg, *s.log, *natsConn, orderService, s.cache)
	orderHandler.Register(s.handler)
	s.log.Info("initialized order routes")

	deliveryStorage := delivery.NewStorage(dbConn, reqTimeout, s.cache)
	deliveryService := delivery.NewService(deliveryStorage, *s.log)
	deliveryHandler := delivery.NewHandler(*s.log, deliveryService, s.cache)
	deliveryHandler.Register(s.handler)
	s.log.Info("initialized delivery routes")

	paymentStorage := payment.NewStorage(dbConn, reqTimeout, s.cache)
	paymentService := payment.NewService(paymentStorage, *s.log)
	paymentHandler := payment.NewHandler(*s.log, paymentService, s.cache)
	paymentHandler.Register(s.handler)
	s.log.Info("initialized payment routes")

	itemStorage := item.NewStorage(dbConn, reqTimeout, s.cache)
	itemService := item.NewService(itemStorage, *s.log)
	itemHandler := item.NewHandler(*s.log, itemService, s.cache)
	itemHandler.Register(s.handler)
	s.log.Info("initialized item routes")

	err := nats2.SubNATS(*s.cfg, *s.log, natsConn, deliveryService, paymentService, itemService, orderService)
	if err != nil {
		log.Fatal("cannot subscribe to NATS:", err)
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
