package handler

import (
	"errors"
	"net/http"

	"coffee-shop-pos/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MenuHandler struct {
	MenuUsecase domain.MenuItemUsecase
}

func NewMenuHandler(u domain.MenuItemUsecase) *MenuHandler {
	return &MenuHandler{
		MenuUsecase: u,
	}
}

func validateMenuItem(item *domain.MenuItem) string {
	if item.Name == "" {
		return "name is required"
	}
	if item.Price.LessThanOrEqual(decimal.Zero) {
		return "price must be greater than zero"
	}
	return ""
}

func (h *MenuHandler) Create(c *gin.Context) {
	var item domain.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateMenuItem(&item); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if err := h.MenuUsecase.Create(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create menu item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *MenuHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	item, err := h.MenuUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menu item"})
		return
	}

	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *MenuHandler) Fetch(c *gin.Context) {
	items, err := h.MenuUsecase.Fetch(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch menu items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *MenuHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var item domain.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateMenuItem(&item); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	item.ID = id
	if err := h.MenuUsecase.Update(c.Request.Context(), &item); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update menu item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	if err := h.MenuUsecase.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete menu item"})
		return
	}

	c.Status(http.StatusNoContent)
}
