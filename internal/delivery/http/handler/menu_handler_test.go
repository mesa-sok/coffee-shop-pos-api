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

type MockMenuItemUsecase struct{ mock.Mock }

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

func (m *MockMenuItemUsecase) Fetch(ctx context.Context, filter domain.MenuFilter) (*domain.MenuListResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuListResult), args.Error(1)
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

		item := domain.MenuItem{Name: "Cappuccino", Price: decimal.NewFromFloat(4.50), Category: "Coffee"}
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

		item := domain.MenuItem{Name: "", Price: decimal.NewFromFloat(4.50)}
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
		item := &domain.MenuItem{ID: id, Name: "Latte", Price: decimal.NewFromFloat(4.00)}
		mockUsecase.On("GetByID", mock.Anything, id).Return(item, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu/"+id.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid id", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu/:id", handler.GetByID)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu/invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUsecase.AssertNotCalled(t, "GetByID")
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

	t.Run("success with filters and pagination", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu", handler.Fetch)

		available := true
		expectedFilter := domain.MenuFilter{Category: "Coffee", IsAvailable: &available, Search: "latte", Limit: 5, Offset: 10}
		items := []domain.MenuItem{{ID: uuid.New(), Name: "Latte", Price: decimal.NewFromFloat(4.00)}}
		mockUsecase.On("Fetch", mock.Anything, mock.MatchedBy(func(filter domain.MenuFilter) bool {
			if filter.IsAvailable == nil || expectedFilter.IsAvailable == nil {
				return false
			}
			return filter.Category == expectedFilter.Category &&
				*filter.IsAvailable == *expectedFilter.IsAvailable &&
				filter.Search == expectedFilter.Search &&
				filter.Limit == expectedFilter.Limit &&
				filter.Offset == expectedFilter.Offset
		})).Return(&domain.MenuListResult{Items: items, Total: 42}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu?category=Coffee&is_available=true&search=latte&limit=5&offset=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response fetchMenuResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 5, response.Limit)
		assert.Equal(t, 10, response.Offset)
		assert.Equal(t, int64(42), response.Total)
		assert.Len(t, response.Data, 1)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with page and page_size", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu", handler.Fetch)

		expected := domain.MenuFilter{Limit: 4, Offset: 8}
		mockUsecase.On("Fetch", mock.Anything, expected).Return(&domain.MenuListResult{Items: []domain.MenuItem{}, Total: 0}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu?page=3&page_size=4", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUsecase.AssertExpectations(t)
	})

	t.Run("invalid is_available", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu", handler.Fetch)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu?is_available=maybe", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUsecase.AssertNotCalled(t, "Fetch")
	})

	t.Run("invalid limit", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu", handler.Fetch)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu?limit=0", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUsecase.AssertNotCalled(t, "Fetch")
	})

	t.Run("usecase error", func(t *testing.T) {
		mockUsecase := new(MockMenuItemUsecase)
		handler := NewMenuHandler(mockUsecase)
		r := gin.Default()
		r.GET("/api/v1/menu", handler.Fetch)

		mockUsecase.On("Fetch", mock.Anything, domain.MenuFilter{Limit: defaultMenuLimit, Offset: 0}).Return(nil, errors.New("db error"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/menu", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
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
		item := domain.MenuItem{Name: "Updated Coffee", Price: decimal.NewFromFloat(5.00), Category: "Coffee"}
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
		item := domain.MenuItem{Name: "Updated Coffee", Price: decimal.NewFromFloat(5.00), Category: "Coffee"}
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
