package usecase

import (
	"context"
	"testing"

	"coffee-shop-pos/internal/domain"
	"github.com/google/uuid"
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
		Price: 3.50,
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.MenuItem")).Return(nil)

	err := u.Create(context.Background(), item)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, item.ID)
	repo.AssertExpectations(t)
}
