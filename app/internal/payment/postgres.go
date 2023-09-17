package payment

import (
	"WBL0/app/internal/apperror"
	"WBL0/app/internal/cache"
	"WBL0/app/internal/model"
	"WBL0/app/pkg/logger"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"time"
)

var _ Storage = &PaymentStorage{}

type PaymentStorage struct {
	log            logger.Logger
	conn           *pgx.Conn
	requestTimeout time.Duration
	cache          *cache.Cache
}

func NewStorage(storage *pgx.Conn, requestTimeout int, cache *cache.Cache) Storage {
	return &PaymentStorage{
		log:            logger.GetLogger(),
		conn:           storage,
		requestTimeout: time.Duration(requestTimeout) * time.Second,
		cache:          cache,
	}
}

func (d *PaymentStorage) Create(payment *Payment) (*Payment, error) {
	d.log.Info("POSTGRES: CREATE PAYMENT")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`INSERT INTO Payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
			 RETURNING id`,
		payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee)

	err := row.Scan(&payment.ID)
	if err != nil {
		err = fmt.Errorf("failed to execute create payment query: %v", err)
		return nil, err
	}

	modelPayment := &model.Payment{
		ID:           payment.ID,
		Transaction:  payment.Transaction,
		RequestID:    payment.RequestID,
		Currency:     payment.Currency,
		Provider:     payment.Provider,
		Amount:       payment.Amount,
		PaymentDt:    payment.PaymentDt,
		Bank:         payment.Bank,
		DeliveryCost: payment.DeliveryCost,
		GoodsTotal:   payment.GoodsTotal,
		CustomFee:    payment.CustomFee,
	}

	d.cache.Payments[payment.ID] = modelPayment

	fmt.Println("\n\nCache after payment creation:")
	for key, value := range d.cache.Payments {
		fmt.Printf("Key: %d, Value: %+v\n", key, value)
	}

	return payment, nil
}

func (d *PaymentStorage) FindById(id int64) (*Payment, error) {
	d.log.Info("POSTGRES: GET PAYMENT BY ID")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`SELECT * FROM Payment
			 WHERE id = $1`, id)
	payment := &Payment{}

	err := row.Scan(
		&payment.ID, &payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
		&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrEmptyString
		}
		err = fmt.Errorf("failed to execute find payment by id query: %v", err)
		return nil, err
	}
	return payment, nil
}
func CacheForPayment(dbConn *pgx.Conn, cache *cache.Cache) error {

	rows, err := dbConn.Query(context.Background(), `SELECT * FROM Payment`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var payment Payment
		err = rows.Scan(
			&payment.ID, &payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)

		if err != nil {
			return err
		}

		modelPayment := &model.Payment{
			ID:           payment.ID,
			Transaction:  payment.Transaction,
			RequestID:    payment.RequestID,
			Currency:     payment.Currency,
			Provider:     payment.Provider,
			Amount:       payment.Amount,
			PaymentDt:    payment.PaymentDt,
			Bank:         payment.Bank,
			DeliveryCost: payment.DeliveryCost,
			GoodsTotal:   payment.GoodsTotal,
			CustomFee:    payment.CustomFee,
		}

		cache.Payments[payment.ID] = modelPayment
	}
	return nil
}