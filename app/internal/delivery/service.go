package delivery

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/pkg/logger"
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, delivery *CreateDeliveryDTO) (*Delivery, error)
	GetById(ctx context.Context, id int64) (*Delivery, error)
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
func (s *service) Create(ctx context.Context, input *CreateDeliveryDTO) (*Delivery, error) {
	s.log.Info("SERVICE: CREATE DELIVERY")

	d := Delivery{
		Name:    input.Name,
		Phone:   input.Phone,
		Zip:     input.Zip,
		City:    input.City,
		Address: input.Address,
		Region:  input.Region,
		Email:   input.Email,
	}

	delivery, err := s.storage.Create(&d)
	if err != nil {
		return nil, err
	}
	return delivery, nil
}

func (s *service) GetById(ctx context.Context, id int64) (*Delivery, error) {
	s.log.Info("SERVICE: GET DELIVERY BY ID")

	delivery, err := s.storage.FindById(id)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			return nil, err
		}
		s.log.Warn("cannot find delivery by id:", err)
		return nil, err
	}
	return delivery, nil
}
