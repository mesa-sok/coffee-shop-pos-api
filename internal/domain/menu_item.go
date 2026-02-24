package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MenuItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Category    string    `json:"category" db:"category"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type MenuItemRepository interface {
	Create(ctx context.Context, item *MenuItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*MenuItem, error)
	Fetch(ctx context.Context) ([]MenuItem, error)
	Update(ctx context.Context, item *MenuItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type MenuItemUsecase interface {
	Create(ctx context.Context, item *MenuItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*MenuItem, error)
	Fetch(ctx context.Context) ([]MenuItem, error)
	Update(ctx context.Context, item *MenuItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}
