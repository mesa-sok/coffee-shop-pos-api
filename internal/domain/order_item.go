package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderItem struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	OrderID    uuid.UUID       `json:"order_id" db:"order_id"`
	MenuItemID uuid.UUID       `json:"menu_item_id" db:"menu_item_id"`
	Quantity   int             `json:"quantity" db:"quantity"`
	UnitPrice  decimal.Decimal `json:"unit_price" db:"unit_price"`
	LineTotal  decimal.Decimal `json:"line_total" db:"line_total"`
}
