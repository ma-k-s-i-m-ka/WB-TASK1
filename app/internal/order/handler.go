package order

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/internal/cache"
	"WBL0/app/internal/handler"
	"WBL0/app/internal/response"
	"WBL0/app/pkg/config"
	"WBL0/app/pkg/logger"
	"encoding/json"
	"errors"
	"github.com/nats-io/nats.go"

	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	orderURL     = "/order"
	orderByIdURL = "/order/:id"
)

type Handler struct {
	cfg          config.Config
	log          logger.Logger
	natsConn     nats.Conn
	orderService Service
	cache        *cache.Cache
}

func NewHandler(cfg config.Config, log logger.Logger, natsConn nats.Conn, orderService Service, cache *cache.Cache) handler.Hand {
	return &Handler{
		cfg:          cfg,
		log:          log,
		natsConn:     natsConn,
		orderService: orderService,
		cache:        cache,
	}
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, orderURL, h.CreateOrder)
	router.HandlerFunc(http.MethodGet, orderByIdURL, h.GetOrderById)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	h.log.Info("HANDLER: CREATE ORDER")

	var input Order

	if err := response.ReadJSON(w, r, &input); err != nil {
		response.BadRequest(w, err.Error(), apperror.ErrInvalidRequestBody.Error())
		return
	}

	orderJSON, err := json.Marshal(input)
	subject := h.cfg.NATS.SUB

	err = h.natsConn.Publish(subject, orderJSON)
	if err != nil {
		response.BadRequest(w, err.Error(), "")
		return
	}

	response.JSON(w, http.StatusOK, "ORDER HAS BEEN SENT")
}

func (h *Handler) GetOrderById(w http.ResponseWriter, r *http.Request) {
	h.log.Infof("HANDLER: GET ORDER BY ID")

	uid := r.URL.Query().Get("uid")
	h.log.Info("Input: ", uid)
	if uid == "" {
		response.BadRequest(w, "empty uid", "")
		return
	}

	cacheOrder, ok := h.cache.Orders[uid]
	if ok {
		h.log.Info("GOT ORDER FROM CACHE BY ID")
		response.JSON(w, http.StatusOK, cacheOrder)
		return
	}

	order, err := h.orderService.GetById(r.Context(), uid)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			response.NotFound(w)
			return
		}
		response.InternalError(w, err.Error(), "")
		return
	}

	h.log.Info("GOT ORDER BY ID")
	response.JSON(w, http.StatusOK, order)
}
