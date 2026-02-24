package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MenuItem struct {
	ID          uuid.UUID       `json:"id" db:"id" binding:"omitempty"`
	Name        string          `json:"name" db:"name" binding:"required"`
	Description string          `json:"description" db:"description"`
	Price       decimal.Decimal `json:"price" db:"price" binding:"required"`
	Category    string          `json:"category" db:"category" binding:"required"`
	IsAvailable bool            `json:"is_available" db:"is_available"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type MenuItemRepository interface {
	Create(ctx context.Context, item *MenuItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*MenuItem, error)
	Fetch(ctx context.Context) ([]MenuItem, error)
	Update(ctx context.Context, item *MenuItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

var ErrNotFound = errors.New("item not found")

type MenuItemUsecase interface {
	Create(ctx context.Context, item *MenuItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*MenuItem, error)
	Fetch(ctx context.Context) ([]MenuItem, error)
	Update(ctx context.Context, item *MenuItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}
