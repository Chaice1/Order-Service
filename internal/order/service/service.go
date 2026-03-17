package Orderservice

import (
	"context"
	"errors"
	"log"

	orderpb "github.com/Chaice1/Order-Service/gen/go/order"
	orderrepo "github.com/Chaice1/Order-Service/internal/order/repository"
)

type OrderEventProducer interface {
	PublishOrderCreated(context.Context, string, string) error
}

type OrderRepo interface {
	CreateOrder(context.Context, string, float64, []*orderrepo.Good) (string, error)
	GetOrderById(context.Context, string) (*orderrepo.GetOrderResponse, error)
}

type OrderService struct {
	repo     OrderRepo
	producer OrderEventProducer
}

func NewOrderService(repo OrderRepo, producer OrderEventProducer) *OrderService {
	return &OrderService{repo: repo, producer: producer}
}

func (os *OrderService) CreateOrder(ctx context.Context, user_id string, goods []*orderpb.Good) (string, error) {

	if len(goods) == 0 {
		return "", errors.New("order must contain at least 1 item")
	}

	items := make([]*orderrepo.Good, len(goods))

	for i, item := range goods {

		items[i] = &orderrepo.Good{
			Id:         item.GetId(),
			NameOfGood: item.GetNameOfGood(),
			Price:      item.GetPrice(),
			Count:      item.GetCount(),
		}
	}

	var total_price float64

	for _, item := range items {
		total_price += (item.Price * float64(item.Count))
	}

	order_id, err := os.repo.CreateOrder(ctx, user_id, total_price, items)

	if err != nil {
		return "", err
	}

	err = os.producer.PublishOrderCreated(ctx, order_id, user_id)

	if err != nil {
		log.Println("couldn't send message")
	}

	return order_id, nil
}

func (os *OrderService) GetOrder(ctx context.Context, order_id string) (*orderrepo.GetOrderResponse, error) {
	order, err := os.repo.GetOrderById(ctx, order_id)
	if err != nil {
		return nil, err
	}
	return order, nil
}
