package item

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

var _ Storage = &ItemStorage{}

type ItemStorage struct {
	log            logger.Logger
	conn           *pgx.Conn
	requestTimeout time.Duration
	cache          *cache.Cache
}

func NewStorage(storage *pgx.Conn, requestTimeout int, cache *cache.Cache) Storage {
	return &ItemStorage{
		log:            logger.GetLogger(),
		conn:           storage,
		requestTimeout: time.Duration(requestTimeout) * time.Second,
		cache:          cache,
	}
}

func (d *ItemStorage) Create(item *Item) (*Item, error) {
	d.log.Info("POSTGRES: CREATE ITEM")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`INSERT INTO Item (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) 
			 RETURNING id`,
		item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)

	err := row.Scan(&item.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute create item query: %v", err)
	}

	modelItem := &model.Item{
		ID:          item.ID,
		ChrtID:      item.ChrtID,
		TrackNumber: item.TrackNumber,
		Price:       item.Price,
		RID:         item.RID,
		Name:        item.Name,
		Sale:        item.Sale,
		Size:        item.Size,
		TotalPrice:  item.TotalPrice,
		NmID:        item.NmID,
		Brand:       item.Brand,
		Status:      item.Status,
	}

	d.cache.Items[item.ID] = modelItem

	fmt.Println("\n\nCache after item creation:")
	for key, value := range d.cache.Items {
		fmt.Printf("Key: %d, Value: %+v\n", key, value)
	}

	return item, nil
}

func (d *ItemStorage) FindById(id int64) (*Item, error) {
	d.log.Info("POSTGRES: GET ITEM BY ID")

	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	row := d.conn.QueryRow(ctx,
		`SELECT * FROM Item
			 WHERE id = $1`, id)
	item := &Item{}

	err := row.Scan(
		&item.ID, &item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
		&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrEmptyString
		}
		err = fmt.Errorf("failed to execute find item by id query: %v", err)
		return nil, err
	}
	return item, nil
}
