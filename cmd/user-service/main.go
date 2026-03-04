package main

import (
	"context"
	"database/sql"
	"io"
	"log"
	"net"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"buf.build/go/protovalidate"
	userpb "github.com/Chaice1/Order-Service/gen/go/user"
	"github.com/Chaice1/Order-Service/internal/user/interceptor"
	userrepo "github.com/Chaice1/Order-Service/internal/user/repository"
	"github.com/Chaice1/Order-Service/internal/user/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type UserService interface {
	Login(context.Context, string, string) (string, error)
	Register(context.Context, string, string) (string, error)
	GetUser(context.Context, string) (*userrepo.User, error)
}
type UserServiceServer struct {
	us UserService
	userpb.UnimplementedUserServiceServer
}

func (uss *UserServiceServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponce, error) {
	id, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	role, ok := ctx.Value("role").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	requested_id := req.GetUserId()
	if id != requested_id && role != "admin" {
		return nil, status.Error(codes.PermissionDenied, codes.Unauthenticated.String())
	}

	user, err := uss.us.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, codes.NotFound.String())
	}

	return &userpb.GetUserResponce{
		UserId:   user.Id,
		Username: user.Username,
		Role:     user.Role,
		IsActive: user.Is_active,
	}, nil
}

func (uss *UserServiceServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponce, error) {
	token, err := uss.us.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, codes.NotFound.String())
	}

	return &userpb.LoginResponce{
		Token: token,
	}, nil
}

func (uss *UserServiceServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponce, error) {
	id, err := uss.us.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, codes.AlreadyExists.String())
	}

	return &userpb.RegisterResponce{
		UserId: id,
	}, nil
}
func RunMigrations(dsn string) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatal(err)
	}

}
func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptor.ValidationInterceptor(validator),
		interceptor.AuthInterceptor(),
	))
	RunMigrations(os.Getenv("DB_DSN"))
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	redisdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

	repo := userrepo.NewRepo(pool, redisdb)
	us := service.NewService(repo)
	repo.CreateAdmin(context.Background(), os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD"))
	UserService := &UserServiceServer{
		us: us,
	}

	userpb.RegisterUserServiceServer(server, UserService)
	log.Println("GRPC server listen port: 8082")
	reflection.Register(server)

	if err := server.Serve(lis); err != io.EOF {
		log.Fatal(err)
	}
}
