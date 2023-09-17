package payment

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
	paymentURL = "/payment/:id"
)

type Handler struct {
	log            logger.Logger
	paymentService Service
	cache          *cache.Cache
}

func NewHandler(log logger.Logger, paymentService Service, cache *cache.Cache) handler.Hand {
	return &Handler{
		log:            log,
		paymentService: paymentService,
		cache:          cache,
	}
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, paymentURL, h.GetPaymentURLById)
}

func (h *Handler) GetPaymentURLById(w http.ResponseWriter, r *http.Request) {
	h.log.Infof("HANDLER: GET PAYMENT BY ID")

	id, err := handler.ReadIdParam64(r)

	h.log.Info("Input: ", id)
	if err != nil {
		response.BadRequest(w, err.Error(), "")
		return
	}

	cachePayment, ok := h.cache.Payments[id]
	if ok {
		h.log.Info("GOT PAYMENT FROM CACHE BY ID")
		response.JSON(w, http.StatusOK, cachePayment)
		return
	}

	payment, err := h.paymentService.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			response.NotFound(w)
			return
		}
		response.InternalError(w, err.Error(), "")
		return
	}

	h.log.Info("GOT PAYMENT BY ID")
	response.JSON(w, http.StatusOK, payment)
}
