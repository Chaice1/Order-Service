package service

import (
	"context"

	"github.com/Chaice1/Order-Service/internal/user/auth"
	userrepo "github.com/Chaice1/Order-Service/internal/user/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceRepository interface {
	GetUserByUsername(context.Context, string) (*userrepo.User, error)
	GetUserById(context.Context, string) (*userrepo.User, error)
	CreateUser(context.Context, string, string) (string, error)
	CreateAdmin(context.Context, string, string) (string, error)
	SetActiveStatus(context.Context, string) error
}

type UserService struct {
	repo UserServiceRepository
}

func NewService(repo UserServiceRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*userrepo.User, error) {
	return s.repo.GetUserById(ctx, id)

}

func (s *UserService) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	err = s.repo.SetActiveStatus(ctx, user.Id)
	if err != nil {
		return "", err
	}

	token, err := auth.Generatetoken(user.Id, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) Register(ctx context.Context, username string, password string) (string, error) {
	id, err := s.repo.CreateUser(ctx, username, password)
	if err != nil {
		return "", err
	}

	return id, nil
}
