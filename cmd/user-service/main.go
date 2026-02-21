package main

import (
	"context"
	"io"
	"log"
	"net"

	"buf.build/go/protovalidate"
	userpb "github.com/Chaice1/Order-Service/gen/go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type UserService struct {
	userpb.UnimplementedUserServiceServer
	validator protovalidate.Validator
}

func (us *UserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponce, error) {

	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}

func main() {

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()

	UserService := &UserService{
		validator: validator,
	}

	userpb.RegisterUserServiceServer(server, UserService)
	log.Println("GRPC server listen port: 8082")
	reflection.Register(server)

	if err := server.Serve(lis); err != io.EOF {
		log.Fatal(err)
	}
}
