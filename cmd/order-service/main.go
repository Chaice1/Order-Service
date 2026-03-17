package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"

	"buf.build/go/protovalidate"
	orderpb "github.com/Chaice1/Order-Service/gen/go/order"
	orderrepo "github.com/Chaice1/Order-Service/internal/order/repository"
	Orderservice "github.com/Chaice1/Order-Service/internal/order/service"
	"github.com/Chaice1/Order-Service/internal/pkg/interceptor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService interface {
	CreateOrder(context.Context, string, []*orderpb.Good) (string, error)
	GetOrder(context.Context, string) (*orderrepo.GetOrderResponse, error)
}
type OrderServiceServer struct {
	orderpb.UnimplementedOrderServiceServer
	os OrderService
}

func (oss *OrderServiceServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	user_id, ok := ctx.Value("user_id").(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	goods := req.GetGoods()
	order_id, err := oss.os.CreateOrder(ctx, user_id, goods)

	if err != nil {
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &orderpb.CreateOrderResponse{
		OrderId: order_id,
		Status:  "CREATED",
	}, nil
}

func (oss *OrderServiceServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	user_id, ok := ctx.Value("user_id").(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	role, ok := ctx.Value("role").(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	order, err := oss.os.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, codes.NotFound.String())
		}
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	if order.OrderInfo.UserId != user_id && role != "admin" {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	items := make([]*orderpb.Good, len(order.Goods))

	for i, item := range order.Goods {
		items[i] = &orderpb.Good{
			Id:         item.Id,
			NameOfGood: item.NameOfGood,
			Price:      item.Price,
			Count:      item.Count,
		}
	}

	return &orderpb.GetOrderResponse{
		OrderId:    order.OrderInfo.Id,
		UserId:     order.OrderInfo.UserId,
		TotalPrice: order.OrderInfo.TotalPrice,
		Status:     order.OrderInfo.Status,
		Goods:      items,
	}, nil

}

func main() {

	godotenv.Load()
	validator, err := protovalidate.New()

	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptor.ValidationInterceptor(validator),
		interceptor.AuthInterceptor(),
	))

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	repo := orderrepo.NewRepo(pool)

	producer := orderrepo.NewOrderProducer(os.Getenv("KAFKA_ADDR"))

	service := Orderservice.NewOrderService(repo, producer)

	OrderServiceServer := &OrderServiceServer{
		os: service,
	}

	orderpb.RegisterOrderServiceServer(server, OrderServiceServer)

	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}

}
