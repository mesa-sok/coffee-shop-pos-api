package handler

import (
	"errors"
	"net/http"

	"coffee-shop-pos/internal/domain"
	"coffee-shop-pos/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	OrderUsecase domain.OrderUsecase
}

type createOrderRequest struct {
	Items []createOrderItemRequest `json:"items"`
}

type createOrderItemRequest struct {
	MenuItemID uuid.UUID `json:"menu_item_id"`
	Quantity   int       `json:"quantity"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func NewOrderHandler(u domain.OrderUsecase) *OrderHandler {
	return &OrderHandler{OrderUsecase: u}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order := &domain.Order{Items: make([]domain.OrderItem, len(req.Items))}
	for i, item := range req.Items {
		order.Items[i] = domain.OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		}
	}

	if err := h.OrderUsecase.Create(c.Request.Context(), order); err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmptyOrderItems), errors.Is(err, usecase.ErrInvalidOrderQuantity):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		}
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	order, err := h.OrderUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	orders, err := h.OrderUsecase.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.OrderUsecase.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		case errors.Is(err, usecase.ErrInvalidOrderStatus), errors.Is(err, usecase.ErrInvalidStatusMove):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
