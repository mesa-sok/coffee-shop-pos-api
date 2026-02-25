package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderRepo struct{ mock.Mock }

type mockMenuRepository struct{ mock.Mock }

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
func (m *mockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}
func (m *mockOrderRepo) List(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, updatedAt time.Time) error {
	args := m.Called(ctx, id, status, updatedAt)
	return args.Error(0)
}

func (m *mockMenuRepository) Create(ctx context.Context, item *domain.MenuItem) error { return nil }
func (m *mockMenuRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}
func (m *mockMenuRepository) Fetch(ctx context.Context) ([]domain.MenuItem, error)    { return nil, nil }
func (m *mockMenuRepository) Update(ctx context.Context, item *domain.MenuItem) error { return nil }
func (m *mockMenuRepository) Delete(ctx context.Context, id uuid.UUID) error          { return nil }

func TestOrderUsecase_Create(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)

	menuID := uuid.New()
	order := &domain.Order{Items: []domain.OrderItem{{MenuItemID: menuID, Quantity: 2}}}
	menuRepo.On("GetByID", mock.Anything, menuID).Return(&domain.MenuItem{ID: menuID, Price: decimal.NewFromFloat(5.50)}, nil)
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	err := u.Create(context.Background(), order)

	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(11).StringFixed(2), order.Subtotal.StringFixed(2))
	assert.Equal(t, decimal.NewFromFloat(1.10).StringFixed(2), order.Tax.StringFixed(2))
	assert.Equal(t, decimal.NewFromFloat(12.10).StringFixed(2), order.Total.StringFixed(2))
	orderRepo.AssertExpectations(t)
	menuRepo.AssertExpectations(t)
}

func TestOrderUsecase_Create_ValidationErrors(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)

	err := u.Create(context.Background(), &domain.Order{})
	assert.ErrorIs(t, err, ErrEmptyOrderItems)

	err = u.Create(context.Background(), &domain.Order{Items: []domain.OrderItem{{MenuItemID: uuid.New(), Quantity: 0}}})
	assert.ErrorIs(t, err, ErrInvalidOrderQuantity)
}

func TestOrderUsecase_UpdateStatus(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)
	id := uuid.New()

	orderRepo.On("GetByID", mock.Anything, id).Return(&domain.Order{ID: id, Status: domain.OrderStatusPending}, nil)
	orderRepo.On("UpdateStatus", mock.Anything, id, domain.OrderStatusPaid, mock.AnythingOfType("time.Time")).Return(nil)

	err := u.UpdateStatus(context.Background(), id, domain.OrderStatusPaid)
	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestOrderUsecase_UpdateStatus_InvalidTransition(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)
	id := uuid.New()

	orderRepo.On("GetByID", mock.Anything, id).Return(&domain.Order{ID: id, Status: domain.OrderStatusPending}, nil)

	err := u.UpdateStatus(context.Background(), id, domain.OrderStatusCompleted)
	assert.ErrorIs(t, err, ErrInvalidStatusMove)
}

func TestOrderUsecase_UpdateStatus_NotFound(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)
	id := uuid.New()

	orderRepo.On("GetByID", mock.Anything, id).Return(nil, nil)

	err := u.UpdateStatus(context.Background(), id, domain.OrderStatusPaid)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOrderUsecase_UpdateStatus_InvalidStatus(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)

	err := u.UpdateStatus(context.Background(), uuid.New(), "unknown")
	assert.ErrorIs(t, err, ErrInvalidOrderStatus)
}

func TestOrderUsecase_UpdateStatus_RepoError(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	menuRepo := new(mockMenuRepository)
	u := NewOrderUsecase(orderRepo, menuRepo)
	id := uuid.New()
	repoErr := errors.New("repo error")

	orderRepo.On("GetByID", mock.Anything, id).Return(&domain.Order{ID: id, Status: domain.OrderStatusPending}, nil)
	orderRepo.On("UpdateStatus", mock.Anything, id, domain.OrderStatusPaid, mock.AnythingOfType("time.Time")).Return(repoErr)

	err := u.UpdateStatus(context.Background(), id, domain.OrderStatusPaid)
	assert.ErrorIs(t, err, repoErr)
}
