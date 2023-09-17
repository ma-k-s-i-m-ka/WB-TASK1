package item

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/pkg/logger"
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, delivery []*CreateItemDTO) ([]*Item, error)
	GetById(ctx context.Context, id int64) (*Item, error)
}

type service struct {
	log     logger.Logger
	storage Storage
}

func NewService(storage Storage, log logger.Logger) Service {
	return &service{
		log:     log,
		storage: storage,
	}
}
func (s *service) Create(ctx context.Context, items []*CreateItemDTO) ([]*Item, error) {
	s.log.Info("SERVICE: CREATE ITEM")

	createdItems := []*Item{}

	for _, input := range items {
		d := Item{
			ChrtID:      input.ChrtID,
			TrackNumber: input.TrackNumber,
			Price:       input.Price,
			RID:         input.RID,
			Name:        input.Name,
			Sale:        input.Sale,
			Size:        input.Size,
			TotalPrice:  input.TotalPrice,
			NmID:        input.NmID,
			Brand:       input.Brand,
			Status:      input.Status,
		}
		item, err := s.storage.Create(&d)
		if err != nil {
			return nil, err
		}
		createdItems = append(createdItems, item)
	}
	return createdItems, nil
}

func (s *service) GetById(ctx context.Context, id int64) (*Item, error) {
	s.log.Info("SERVICE: GET ITEM BY ID")

	item, err := s.storage.FindById(id)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			return nil, err
		}
		s.log.Warn("cannot find item by id:", err)
		return nil, err
	}
	return item, nil
}
