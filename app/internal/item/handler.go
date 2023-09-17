package item

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
	itemURL = "/item/:id"
)

type Handler struct {
	log         logger.Logger
	itemService Service
	cache       *cache.Cache
}

func NewHandler(log logger.Logger, itemService Service, cache *cache.Cache) handler.Hand {
	return &Handler{
		log:         log,
		itemService: itemService,
		cache:       cache,
	}
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, itemURL, h.GetItemById)
}

func (h *Handler) GetItemById(w http.ResponseWriter, r *http.Request) {
	h.log.Info("HANDLER: GET ITEM BY ID")

	id, err := handler.ReadIdParam64(r)

	h.log.Info("Input: ", id)
	if err != nil {
		response.BadRequest(w, err.Error(), "")
		return
	}

	cacheItem, ok := h.cache.Items[id]
	if ok {
		h.log.Info("GOT ITEM FROM CACHE BY ID")
		response.JSON(w, http.StatusOK, cacheItem)
		return
	}

	item, err := h.itemService.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			response.NotFound(w)
			return
		}
		response.InternalError(w, err.Error(), "")
		return
	}

	h.log.Info("GOT ITEM BY ID")
	response.JSON(w, http.StatusOK, item)
}
