package payment

type Storage interface {
	Create(payment *Payment) (*Payment, error)
	FindById(id int64) (*Payment, error)
}
