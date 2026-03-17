package main

import (
	"context"
	"log"
	"net/http"

	orderpb "github.com/Chaice1/Order-Service/gen/go/order"
	userpb "github.com/Chaice1/Order-Service/gen/go/user"
	"github.com/Chaice1/Order-Service/internal/api-gateway/middleware"
	apierrors "github.com/Chaice1/Order-Service/internal/pkg/api-errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GatewayService struct {
	userpb.UserServiceClient
	orderpb.OrderServiceClient
}

type UserRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GetUserRequestBody struct {
	User_id string `json:"user_id"`
}

type GetOrderRequestBody struct {
	Order_id string `json:"order_id"`
}

type Good struct {
	Id         string  `json:"good_id"`
	NameOfGood string  `json:"name_of_good"`
	Count      int64   `json:"count"`
	Price      float64 `json:"price"`
}

type CreateOrderRequestBody struct {
	Goods []*Good `json:"goods"`
}

func HelperForCreatinMetadata(c *gin.Context) context.Context {
	token := c.GetString("token")

	ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.Pairs("authorization", token))
	return ctx
}

func (gs *GatewayService) Register(c *gin.Context) {
	var user UserRequestBody

	if err := c.ShouldBindJSON(&user); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	responce, err := gs.UserServiceClient.Register(c.Request.Context(), &userpb.RegisterRequest{
		Username: user.Username,
		Password: user.Password,
	})

	if err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user was registered successfully",
		"user_id": responce.UserId,
	})

}

func (gc *GatewayService) Login(c *gin.Context) {
	var user UserRequestBody

	if err := c.ShouldBindJSON(&user); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	responce, err := gc.UserServiceClient.Login(c.Request.Context(), &userpb.LoginRequest{
		Username: user.Username,
		Password: user.Password,
	})

	if err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user logined successfully",
		"token":   responce.Token,
	})

}

func (gs *GatewayService) GetUser(c *gin.Context) {
	var reqbody GetUserRequestBody

	if err := c.ShouldBindJSON(&reqbody); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	if _, err := uuid.Parse(reqbody.User_id); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}
	ctx := HelperForCreatinMetadata(c)
	responce, err := gs.UserServiceClient.GetUser(ctx, &userpb.GetUserRequest{
		UserId: reqbody.User_id,
	})

	if err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "user is found",
		"UserId":   responce.UserId,
		"Username": responce.Username,
		"Role":     responce.Role,
		"IsActive": responce.IsActive,
	})

}

func (gc *GatewayService) GetOrder(c *gin.Context) {
	var req GetOrderRequestBody

	if err := c.BindJSON(&req); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	if _, err := uuid.Parse(req.Order_id); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	ctx := HelperForCreatinMetadata(c)
	resp, err := gc.OrderServiceClient.GetOrder(ctx, &orderpb.GetOrderRequest{
		OrderId: req.Order_id,
	})

	if err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "order was found",
		"data":    resp,
	})
}

func (gs *GatewayService) CreateOrder(c *gin.Context) {
	var req CreateOrderRequestBody

	if err := c.BindJSON(&req); err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	goods := make([]*orderpb.Good, len(req.Goods))

	for i, good := range req.Goods {
		goods[i] = &orderpb.Good{
			Id:         good.Id,
			NameOfGood: good.NameOfGood,
			Count:      good.Count,
			Price:      good.Price,
		}
	}
	ctx := HelperForCreatinMetadata(c)
	resp, err := gs.OrderServiceClient.CreateOrder(ctx, &orderpb.CreateOrderRequest{
		Goods: goods,
	})

	if err != nil {
		apierrors.HandleErrors(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "order was created",
		"Order_id": resp.OrderId,
		"Status":   resp.Status,
	})

}

func main() {
	server := gin.New()

	conn1, err := grpc.NewClient("user_service:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	conn2, err := grpc.NewClient("order_service:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client1 := userpb.NewUserServiceClient(conn1)
	client2 := orderpb.NewOrderServiceClient(conn2)
	gs := GatewayService{client1, client2}

	server.POST("/login", gs.Login)
	server.POST("/register", gs.Register)

	ApiWithAuth := server.Group("/auth")

	ApiWithAuth.Use(middleware.AuthMiddleware())

	ApiWithAuth.GET("/get_user", gs.GetUser)
	ApiWithAuth.POST("/create_order", gs.CreateOrder)
	ApiWithAuth.GET("/get_order", gs.GetOrder)

	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
