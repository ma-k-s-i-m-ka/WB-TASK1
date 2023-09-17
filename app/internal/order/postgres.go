package order

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

var _ Storage = &OrderStorage{}

type OrderStorage struct {
	log            logger.Logger
	conn           *pgx.Conn
	requestTimeout time.Duration
	cache          *cache.Cache
}

func NewStorage(storage *pgx.Conn, requestTimeout int, cache *cache.Cache) Storage {
	return &OrderStorage{
		log:            logger.GetLogger(),
		conn:           storage,
		requestTimeout: time.Duration(requestTimeout) * time.Second,
		cache:          cache,
	}
}

func (d *OrderStorage) Create(order *CreateOrderDTO) (*CreateOrderDTO, error) {
	d.log.Info("POSTGRES: CREATE ORDER")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	_, err := d.conn.Exec(ctx,
		`INSERT INTO "order" (order_uid, track_number, entry, delivery, payment, items, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
         VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Delivery, order.Payment, order.Items, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OofShard)

	if err != nil {
		return nil, fmt.Errorf("failed to execute create order query: %v", err)
	}

	modelOrder := &model.Order{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Delivery:          order.Delivery,
		Payment:           order.Payment,
		Items:             order.Items,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		ShardKey:          order.ShardKey,
		SMID:              order.SMID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}

	d.cache.Orders[order.OrderUID] = modelOrder

	fmt.Println("\n\nCache after order creation:")
	for key, value := range d.cache.Orders {
		fmt.Printf("Key: %s, Value: %+v\n", key, value)
	}

	return order, nil
}

func (d *OrderStorage) FindById(uid string) (*CreateOrderDTO, error) {
	d.log.Info("POSTGRES: GET ORDER BY ID")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`SELECT * FROM "order"
			 WHERE order_uid = $1`, uid)
	order := &CreateOrderDTO{}

	err := row.Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Delivery, &order.Payment, &order.Items,
		&order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.ShardKey, &order.SMID, &order.DateCreated, &order.OofShard)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrEmptyString
		}
		err = fmt.Errorf("failed to execute find order by id query: %v", err)
		return nil, err
	}
	return order, nil
}

func CacheForOrder(dbConn *pgx.Conn, cache *cache.Cache) error {

	rows, err := dbConn.Query(context.Background(), `SELECT * FROM "order"`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order CreateOrderDTO
		err = rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Delivery, &order.Payment, &order.Items,
			&order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.ShardKey, &order.SMID, &order.DateCreated, &order.OofShard)
		if err != nil {
			return err
		}
		modelOrder := &model.Order{
			OrderUID:          order.OrderUID,
			TrackNumber:       order.TrackNumber,
			Entry:             order.Entry,
			Delivery:          order.Delivery,
			Payment:           order.Payment,
			Items:             order.Items,
			Locale:            order.Locale,
			InternalSignature: order.InternalSignature,
			CustomerID:        order.CustomerID,
			DeliveryService:   order.DeliveryService,
			ShardKey:          order.ShardKey,
			SMID:              order.SMID,
			DateCreated:       order.DateCreated,
			OofShard:          order.OofShard,
		}
		cache.Orders[order.OrderUID] = modelOrder
	}
	return nil
}
