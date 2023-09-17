package delivery

type Storage interface {
	Create(delivery *Delivery) (*Delivery, error)
	FindById(id int64) (*Delivery, error)
}
