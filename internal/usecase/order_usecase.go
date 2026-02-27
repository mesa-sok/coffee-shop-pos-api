package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrEmptyOrderItems      = errors.New("order must contain at least one item")
	ErrInvalidOrderQuantity = errors.New("quantity must be greater than zero")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrInvalidStatusMove    = errors.New("invalid status transition")
)

var allowedStatusTransitions = map[string]map[string]bool{
	domain.OrderStatusPending: {
		domain.OrderStatusPaid:      true,
		domain.OrderStatusCancelled: true,
	},
	domain.OrderStatusPaid: {
		domain.OrderStatusCompleted: true,
		domain.OrderStatusCancelled: true,
	},
	domain.OrderStatusCancelled: {},
	domain.OrderStatusCompleted: {},
}

type orderUsecase struct {
	orderRepo domain.OrderRepository
	menuRepo  domain.MenuItemRepository
	taxRate   decimal.Decimal
}

func NewOrderUsecase(orderRepo domain.OrderRepository, menuRepo domain.MenuItemRepository) domain.OrderUsecase {
	return &orderUsecase{
		orderRepo: orderRepo,
		menuRepo:  menuRepo,
		taxRate:   decimal.NewFromFloat(0.10),
	}
}

func (u *orderUsecase) Create(ctx context.Context, order *domain.Order) error {
	if len(order.Items) == 0 {
		return ErrEmptyOrderItems
	}

	now := time.Now()
	order.ID = uuid.New()
	order.OrderNumber = fmt.Sprintf("ORD-%d", now.UnixNano())
	order.Status = domain.OrderStatusPending
	order.CreatedAt = now
	order.UpdatedAt = now

	subtotal := decimal.Zero
	for i := range order.Items {
		if order.Items[i].Quantity <= 0 {
			return ErrInvalidOrderQuantity
		}

		menuItem, err := u.menuRepo.GetByID(ctx, order.Items[i].MenuItemID)
		if err != nil {
			return err
		}
		if menuItem == nil {
			return domain.ErrNotFound
		}

		lineTotal := menuItem.Price.Mul(decimal.NewFromInt(int64(order.Items[i].Quantity)))
		order.Items[i].ID = uuid.New()
		order.Items[i].OrderID = order.ID
		order.Items[i].UnitPrice = menuItem.Price
		order.Items[i].LineTotal = lineTotal
		subtotal = subtotal.Add(lineTotal)
	}

	order.Subtotal = subtotal.Round(2)
	order.Tax = order.Subtotal.Mul(u.taxRate).Round(2)
	order.Total = order.Subtotal.Add(order.Tax).Round(2)

	return u.orderRepo.Create(ctx, order)
}

func (u *orderUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return u.orderRepo.GetByID(ctx, id)
}

func (u *orderUsecase) List(ctx context.Context) ([]domain.Order, error) {
	return u.orderRepo.List(ctx)
}

func (u *orderUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if _, ok := allowedStatusTransitions[status]; !ok {
		return ErrInvalidOrderStatus
	}

	order, err := u.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return domain.ErrNotFound
	}

	if order.Status == status {
		return nil
	}

	if !allowedStatusTransitions[order.Status][status] {
		return ErrInvalidStatusMove
	}

	err = u.orderRepo.UpdateStatus(ctx, id, status, time.Now())
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
