package payment

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/pkg/logger"
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, delivery *CreatePaymentDTO) (*Payment, error)
	GetById(ctx context.Context, id int64) (*Payment, error)
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
func (s *service) Create(ctx context.Context, input *CreatePaymentDTO) (*Payment, error) {
	s.log.Info("SERVICE: CREATE PAYMENT")

	d := Payment{
		Transaction:  input.Transaction,
		RequestID:    input.RequestID,
		Currency:     input.Currency,
		Provider:     input.Provider,
		Amount:       input.Amount,
		PaymentDt:    input.PaymentDt,
		Bank:         input.Bank,
		DeliveryCost: input.DeliveryCost,
		GoodsTotal:   input.GoodsTotal,
		CustomFee:    input.CustomFee,
	}

	payment, err := s.storage.Create(&d)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *service) GetById(ctx context.Context, id int64) (*Payment, error) {
	s.log.Info("SERVICE: GET PAYMENT BY ID")

	payment, err := s.storage.FindById(id)
	if err != nil {
		if errors.Is(err, apperror.ErrEmptyString) {
			return nil, err
		}
		s.log.Warn("cannot find payment by id:", err)
		return nil, err
	}
	return payment, nil
}
