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
	"coffee-shop-pos/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderUsecase struct{ mock.Mock }

func (m *mockOrderUsecase) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
func (m *mockOrderUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}
func (m *mockOrderUsecase) List(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *mockOrderUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestOrderHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.POST("/api/v1/orders", h.Create)

	menuID := uuid.New()
	payload := map[string]any{"items": []map[string]any{{"menu_item_id": menuID, "quantity": 2}}}
	body, _ := json.Marshal(payload)
	mockUsecase.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(nil)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.GET("/api/v1/orders/:id", h.GetByID)

	id := uuid.New()
	mockUsecase.On("GetByID", mock.Anything, id).Return(&domain.Order{ID: id}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.GET("/api/v1/orders", h.List)

	mockUsecase.On("List", mock.Anything).Return([]domain.Order{{ID: uuid.New()}}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_UpdateStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.PATCH("/api/v1/orders/:id/status", h.UpdateStatus)

	id := uuid.New()
	body, _ := json.Marshal(map[string]string{"status": domain.OrderStatusPaid})
	mockUsecase.On("UpdateStatus", mock.Anything, id, domain.OrderStatusPaid).Return(nil)

	req, _ := http.NewRequest(http.MethodPatch, "/api/v1/orders/"+id.String()+"/status", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestOrderHandler_UpdateStatus_InvalidTransition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.PATCH("/api/v1/orders/:id/status", h.UpdateStatus)

	id := uuid.New()
	body, _ := json.Marshal(map[string]string{"status": domain.OrderStatusCompleted})
	mockUsecase.On("UpdateStatus", mock.Anything, id, domain.OrderStatusCompleted).Return(usecase.ErrInvalidStatusMove)

	req, _ := http.NewRequest(http.MethodPatch, "/api/v1/orders/"+id.String()+"/status", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderHandler_Create_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.POST("/api/v1/orders", h.Create)

	menuID := uuid.New()
	payload := map[string]any{"items": []map[string]any{{"menu_item_id": menuID, "quantity": 1}}}
	body, _ := json.Marshal(payload)
	mockUsecase.On("Create", mock.Anything, mock.AnythingOfType("*domain.Order")).Return(domain.ErrNotFound)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderHandler_List_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUsecase := new(mockOrderUsecase)
	h := NewOrderHandler(mockUsecase)
	r := gin.Default()
	r.GET("/api/v1/orders", h.List)

	mockUsecase.On("List", mock.Anything).Return(nil, errors.New("db error"))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
