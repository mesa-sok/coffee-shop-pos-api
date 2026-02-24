package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-shop-pos/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMenuItemUsecase is a mock implementation of domain.MenuItemUsecase
type MockMenuItemUsecase struct {
	mock.Mock
}

func (m *MockMenuItemUsecase) Create(ctx context.Context, item *domain.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockMenuItemUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}

func (m *MockMenuItemUsecase) Fetch(ctx context.Context) ([]domain.MenuItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MenuItem), args.Error(1)
}

func (m *MockMenuItemUsecase) Update(ctx context.Context, item *domain.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockMenuItemUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestMenuHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.POST("/api/v1/menu", handler.Create)

		item := domain.MenuItem{
			Name:     "Cappuccino",
			Price:    decimal.NewFromFloat(4.50),
			Category: "Coffee",
		}

		mockUsecase.On("Create", mock.Anything, mock.MatchedBy(func(i *domain.MenuItem) bool {
			return i.Name == item.Name && i.Price.Equal(item.Price)
		})).Return(nil)

		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/menu", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid input", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.POST("/api/v1/menu", handler.Create)

		item := domain.MenuItem{
			Name:  "", // Invalid: empty name
			Price: decimal.NewFromFloat(4.50),
		}

		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/menu", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUsecase.AssertNotCalled(t, "Create")
	})
}

func TestMenuHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu/:id", handler.GetByID)

		id := uuid.New()
		item := &domain.MenuItem{
			ID:    id,
			Name:  "Latte",
			Price: decimal.NewFromFloat(4.00),
		}

		mockUsecase.On("GetByID", mock.Anything, id).Return(item, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.MenuItem
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, item.ID, response.ID)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu/:id", handler.GetByID)

		id := uuid.New()
		mockUsecase.On("GetByID", mock.Anything, id).Return(nil, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUsecase.AssertExpectations(t)
	})
}

func TestMenuHandler_Fetch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu", handler.Fetch)

		items := []domain.MenuItem{
			{ID: uuid.New(), Name: "Espresso", Price: decimal.NewFromFloat(2.50)},
			{ID: uuid.New(), Name: "Tea", Price: decimal.NewFromFloat(2.00)},
		}

		mockUsecase.On("Fetch", mock.Anything).Return(items, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.MenuItem
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		mockUsecase.AssertExpectations(t)
	})
}

func TestMenuHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.PUT("/api/v1/menu/:id", handler.Update)

		id := uuid.New()
		item := domain.MenuItem{
			Name:     "Updated Coffee",
			Price:    decimal.NewFromFloat(5.00),
			Category: "Coffee",
		}

		mockUsecase.On("Update", mock.Anything, mock.MatchedBy(func(i *domain.MenuItem) bool {
			return i.ID == id && i.Name == item.Name
		})).Return(nil)

		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/menu/"+id.String(), bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.PUT("/api/v1/menu/:id", handler.Update)

		id := uuid.New()
		item := domain.MenuItem{
			Name:     "Updated Coffee",
			Price:    decimal.NewFromFloat(5.00),
			Category: "Coffee",
		}

		mockUsecase.On("Update", mock.Anything, mock.Anything).Return(domain.ErrNotFound)

		body, _ := json.Marshal(item)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/menu/"+id.String(), bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUsecase.AssertExpectations(t)
	})
}

func TestMenuHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.DELETE("/api/v1/menu/:id", handler.Delete)

		id := uuid.New()
		mockUsecase.On("Delete", mock.Anything, id).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/menu/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.DELETE("/api/v1/menu/:id", handler.Delete)

		id := uuid.New()
		mockUsecase.On("Delete", mock.Anything, id).Return(errors.New("db error"))

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/menu/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUsecase.AssertExpectations(t)
	})
}
