package order

type Storage interface {
	Create(order *CreateOrderDTO) (*CreateOrderDTO, error)
	FindById(uid string) (*CreateOrderDTO, error)
}
