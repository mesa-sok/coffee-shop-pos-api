package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"
	OrderStatusCompleted = "completed"
)

type Order struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	OrderNumber string          `json:"order_number" db:"order_number"`
	Status      string          `json:"status" db:"status"`
	Subtotal    decimal.Decimal `json:"subtotal" db:"subtotal"`
	Tax         decimal.Decimal `json:"tax" db:"tax"`
	Total       decimal.Decimal `json:"total" db:"total"`
	Items       []OrderItem     `json:"items,omitempty"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	List(ctx context.Context) ([]Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) error
}

type OrderUsecase interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	List(ctx context.Context) ([]Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}
