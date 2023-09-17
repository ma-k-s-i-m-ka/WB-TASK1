package item

type Storage interface {
	Create(item *Item) (*Item, error)
	FindById(id int64) (*Item, error)
}
