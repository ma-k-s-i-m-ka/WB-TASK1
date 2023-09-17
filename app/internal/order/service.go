package order

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/pkg/logger"
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, order *CreateOrderDTO) (*CreateOrderDTO, error)
	GetById(ctx context.Context, uid string) (*CreateOrderDTO, error)
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
func (s *service) Create(ctx context.Context, input *CreateOrderDTO) (*CreateOrderDTO, error) {
	s.log.Info("SERVICE: CREATE ORDER")

	o := CreateOrderDTO{
		OrderUID:          input.OrderUID,
		TrackNumber:       input.TrackNumber,
		Entry:             input.Entry,
		Delivery:          input.Delivery,
		Payment:           input.Payment,
		Items:             input.Items,
		Locale:            input.Locale,
		InternalSignature: input.InternalSignature,
		CustomerID:        input.CustomerID,
		DeliveryService:   input.DeliveryService,
		ShardKey:          input.ShardKey,
		SMID:              input.SMID,
		DateCreated:       input.DateCreated,
		OofShard:          input.OofShard,
	}

	order, err := s.storage.Create(&o)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *service) GetById(ctx context.Context, uid string) (*CreateOrderDTO, error) {
	s.log.Info("SERVICE: GET ORDER BY ID")

	order, err := s.storage.FindById(uid)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			return nil, err
		}
		s.log.Warn("cannot find order by id:", err)
		return nil, err
	}
	return order, nil
}
