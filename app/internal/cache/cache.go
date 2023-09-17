package cache

import (
	"WBL0/app/internal/model"
)

type Cache struct {
	Deliveries map[int64]*model.Delivery
	Payments   map[int64]*model.Payment
	Items      map[int64]*model.Item
	Orders     map[string]*model.Order
}
