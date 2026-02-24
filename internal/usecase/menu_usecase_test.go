package usecase

import (
	"context"
	"testing"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Manual mock for repository
type mockMenuRepo struct {
	mock.Mock
}

func (m *mockMenuRepo) Create(ctx context.Context, item *domain.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockMenuRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}

func (m *mockMenuRepo) Fetch(ctx context.Context) ([]domain.MenuItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.MenuItem), args.Error(1)
}

func (m *mockMenuRepo) Update(ctx context.Context, item *domain.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockMenuRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreate(t *testing.T) {
	repo := new(mockMenuRepo)
	u := NewMenuUsecase(repo)

	item := &domain.MenuItem{
		Name:  "Test Coffee",
		Price: decimal.NewFromFloat(3.50),
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MenuItem")).Return(nil)

	err := u.Create(context.Background(), item)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, item.ID)
	repo.AssertExpectations(t)
}

func TestGetByID(t *testing.T) {
	repo := new(mockMenuRepo)
	u := NewMenuUsecase(repo)
	id := uuid.New()
	expected := &domain.MenuItem{
		ID:    id,
		Name:  "Latte",
		Price: decimal.NewFromFloat(4.25),
	}

	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := u.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestFetch(t *testing.T) {
	repo := new(mockMenuRepo)
	u := NewMenuUsecase(repo)
	items := []domain.MenuItem{
		{
			ID:    uuid.New(),
			Name:  "Espresso",
			Price: decimal.NewFromFloat(2.50),
		},
		{
			ID:    uuid.New(),
			Name:  "Cappuccino",
			Price: decimal.NewFromFloat(3.75),
		},
	}

	repo.On("Fetch", mock.Anything).Return(items, nil)

	result, err := u.Fetch(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, len(items))
	assert.Equal(t, items, result)
	repo.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	repo := new(mockMenuRepo)
	u := NewMenuUsecase(repo)
	id := uuid.New()
	item := &domain.MenuItem{
		ID:    id,
		Name:  "Mocha",
		Price: decimal.NewFromFloat(4.00),
	}
	existing := &domain.MenuItem{
		ID: id,
	}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, item).Return(nil)

	err := u.Update(context.Background(), item)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	repo := new(mockMenuRepo)
	u := NewMenuUsecase(repo)
	id := uuid.New()

	repo.On("Delete", mock.Anything, id).Return(nil)

	err := u.Delete(context.Background(), id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
