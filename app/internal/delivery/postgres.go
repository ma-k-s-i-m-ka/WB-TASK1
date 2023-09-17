package delivery

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

var _ Storage = &DeliveryStorage{}

type DeliveryStorage struct {
	log            logger.Logger
	conn           *pgx.Conn
	requestTimeout time.Duration
	cache          *cache.Cache
}

func NewStorage(storage *pgx.Conn, requestTimeout int, cache *cache.Cache) Storage {
	return &DeliveryStorage{
		log:            logger.GetLogger(),
		conn:           storage,
		requestTimeout: time.Duration(requestTimeout) * time.Second,
		cache:          cache,
	}
}

func (d *DeliveryStorage) Create(delivery *Delivery) (*Delivery, error) {
	d.log.Info("POSTGRES: CREATE DELIVERY")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`INSERT INTO Delivery (name, phone, zip, city, address, region, email)
			 VALUES($1,$2,$3,$4,$5,$6,$7) 
			 RETURNING id`,
		delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email)

	err := row.Scan(&delivery.ID)
	if err != nil {
		err = fmt.Errorf("failed to execute create delivery query: %v", err)
		return nil, err
	}

	modelDelivery := &model.Delivery{
		ID:      delivery.ID,
		Name:    delivery.Name,
		Phone:   delivery.Phone,
		Zip:     delivery.Zip,
		City:    delivery.City,
		Address: delivery.Address,
		Region:  delivery.Region,
		Email:   delivery.Email,
	}

	d.cache.Deliveries[delivery.ID] = modelDelivery

	fmt.Println("\n\nCache after delivery creation:")
	for key, value := range d.cache.Deliveries {
		fmt.Printf("Key: %d, Value: %+v\n", key, value)
	}

	return delivery, nil
}

func (d *DeliveryStorage) FindById(id int64) (*Delivery, error) {
	d.log.Info("POSTGRES: GET DELIVERY BY ID")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`SELECT * FROM Delivery
			 WHERE id = $1`, id)
	delivery := &Delivery{}

	err := row.Scan(
		&delivery.ID, &delivery.Name, &delivery.Phone, &delivery.Zip,
		&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrEmptyString
		}
		err = fmt.Errorf("failed to execute find delivery by id query: %v", err)
		return nil, err
	}
	return delivery, nil
}
