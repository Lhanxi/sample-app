package item

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidInput = errors.New("invalid item input")

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Item, error) {
	return s.repository.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (Item, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateItemRequest) (Item, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" {
		return Item{}, ErrInvalidInput
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) Update(
	ctx context.Context,
	id string,
	input UpdateItemRequest,
) (Item, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" {
		return Item{}, ErrInvalidInput
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}
