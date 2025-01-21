package userservice

import (
	"context"
	"github.com/yuisofull/gommunigate/internal/usersvc/pkg/model"
)

type Service interface {
	CreateProfile(ctx context.Context, u model.User) error
	GetProfile(ctx context.Context, uid string, authenticated bool) (model.User, error)
	UpdateProfile(ctx context.Context, u model.User) error
	DeleteProfile(ctx context.Context, uid string) error
}

type Repository interface {
	CreateUser(ctx context.Context, u model.User) error
	GetUser(ctx context.Context, uid string) (model.User, error)
	UpdateUser(ctx context.Context, u model.User) error
	DeleteUser(ctx context.Context, uid string) error
}

func NewService(r Repository) Service {
	if r == nil {
		panic("invalid repository")
	}
	return service{repo: r}
}

type service struct {
	repo Repository
}

func (s service) CreateProfile(ctx context.Context, u model.User) error {
	return s.repo.CreateUser(ctx, u)
}

func (s service) GetProfile(ctx context.Context, uid string, authenticated bool) (model.User, error) {
	return s.repo.GetUser(ctx, uid)
}

func (s service) UpdateProfile(ctx context.Context, u model.User) error {
	return s.repo.UpdateUser(ctx, u)
}

func (s service) DeleteProfile(ctx context.Context, uid string) error {
	return s.repo.DeleteUser(ctx, uid)
}
