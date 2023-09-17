package delivery

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/internal/cache"
	"WBL0/app/internal/handler"
	"WBL0/app/internal/response"
	"WBL0/app/pkg/logger"
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	deliveryURL = "/delivery/:id"
)

type Handler struct {
	log             logger.Logger
	deliveryService Service
	cache           *cache.Cache
}

func NewHandler(log logger.Logger, deliveryService Service, cache *cache.Cache) handler.Hand {
	return &Handler{
		log:             log,
		deliveryService: deliveryService,
		cache:           cache,
	}
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, deliveryURL, h.GetDeliveryById)
}

func (h *Handler) GetDeliveryById(w http.ResponseWriter, r *http.Request) {
	h.log.Infof("HANDLER: GET DELIVERY BY ID")

	id, err := handler.ReadIdParam64(r)

	h.log.Info("Input: ", id)
	if err != nil {
		response.BadRequest(w, err.Error(), "")
		return
	}

	cacheDelivery, ok := h.cache.Deliveries[id]
	if ok {
		h.log.Info("GOT DELIVERY FROM CACHE BY ID")
		response.JSON(w, http.StatusOK, cacheDelivery)
		return
	}

	delivery, err := h.deliveryService.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			response.NotFound(w)
			return
		}
		response.InternalError(w, err.Error(), "")
		return
	}

	h.log.Info("GOT DELIVERY BY ID")
	response.JSON(w, http.StatusOK, delivery)
}
