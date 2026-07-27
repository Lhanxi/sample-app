package item

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	items       []Item
	item        Item
	err         error
	createInput CreateItemRequest
	updateInput UpdateItemRequest
}

func (f *fakeRepository) List(context.Context) ([]Item, error) {
	return f.items, f.err
}

func (f *fakeRepository) GetByID(context.Context, string) (Item, error) {
	return f.item, f.err
}

func (f *fakeRepository) Create(
	_ context.Context,
	input CreateItemRequest,
) (Item, error) {
	f.createInput = input
	return f.item, f.err
}

func (f *fakeRepository) Update(
	_ context.Context,
	_ string,
	input UpdateItemRequest,
) (Item, error) {
	f.updateInput = input
	return f.item, f.err
}

func (f *fakeRepository) Delete(context.Context, string) error {
	return f.err
}

func TestServiceCreate(t *testing.T) {
	t.Run("creates valid item and trims input", func(t *testing.T) {
		repository := &fakeRepository{
			item: Item{ID: "item-id", Name: "Desk"},
		}
		service := NewService(repository)

		created, err := service.Create(context.Background(), CreateItemRequest{
			Name:        "  Desk  ",
			Description: "  Standing desk  ",
		})

		if err != nil {
			t.Fatalf("Create() error = %v; want nil", err)
		}
		if created.ID != "item-id" {
			t.Errorf("Create() ID = %q; want %q", created.ID, "item-id")
		}
		if repository.createInput.Name != "Desk" {
			t.Errorf("repository name = %q; want %q", repository.createInput.Name, "Desk")
		}
		if repository.createInput.Description != "Standing desk" {
			t.Errorf(
				"repository description = %q; want %q",
				repository.createInput.Description,
				"Standing desk",
			)
		}
	})

	t.Run("rejects blank name", func(t *testing.T) {
		service := NewService(&fakeRepository{})

		_, err := service.Create(context.Background(), CreateItemRequest{Name: "   "})

		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("Create() error = %v; want %v", err, ErrInvalidInput)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		expectedError := errors.New("database failed")
		service := NewService(&fakeRepository{err: expectedError})

		_, err := service.Create(context.Background(), CreateItemRequest{Name: "Desk"})

		if !errors.Is(err, expectedError) {
			t.Fatalf("Create() error = %v; want %v", err, expectedError)
		}
	})
}

func TestServiceUpdateValidatesAndTrimsInput(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.Update(context.Background(), "item-id", UpdateItemRequest{
		Name:        "  Chair  ",
		Description: "  Office chair  ",
	})

	if err != nil {
		t.Fatalf("Update() error = %v; want nil", err)
	}
	if repository.updateInput.Name != "Chair" {
		t.Errorf("repository name = %q; want %q", repository.updateInput.Name, "Chair")
	}
	if repository.updateInput.Description != "Office chair" {
		t.Errorf(
			"repository description = %q; want %q",
			repository.updateInput.Description,
			"Office chair",
		)
	}

	_, err = service.Update(
		context.Background(),
		"item-id",
		UpdateItemRequest{Name: " "},
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Update() blank name error = %v; want %v", err, ErrInvalidInput)
	}
}

func TestServicePropagatesRepositoryErrors(t *testing.T) {
	expectedError := ErrNotFound
	service := NewService(&fakeRepository{err: expectedError})

	if _, err := service.List(context.Background()); !errors.Is(err, expectedError) {
		t.Errorf("List() error = %v; want %v", err, expectedError)
	}
	if _, err := service.GetByID(context.Background(), "id"); !errors.Is(err, expectedError) {
		t.Errorf("GetByID() error = %v; want %v", err, expectedError)
	}
	if err := service.Delete(context.Background(), "id"); !errors.Is(err, expectedError) {
		t.Errorf("Delete() error = %v; want %v", err, expectedError)
	}
}
