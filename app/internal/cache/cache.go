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

func NewCache() *Cache {
	return &Cache{
		Deliveries: make(map[int64]*model.Delivery),
		Payments:   make(map[int64]*model.Payment),
		Items:      make(map[int64]*model.Item),
		Orders:     make(map[string]*model.Order),
	}
}
