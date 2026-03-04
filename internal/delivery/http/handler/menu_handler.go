package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"coffee-shop-pos/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	defaultMenuLimit = 20
	maxMenuLimit     = 100
)

type fetchMenuRequest struct {
	Category    string
	IsAvailable *bool
	Search      string
	Limit       int
	Offset      int
}

type fetchMenuResponse struct {
	Data   []domain.MenuItem `json:"data"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
	Total  int64             `json:"total"`
}

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

func (h *MenuHandler) parseFetchMenuRequest(c *gin.Context) (*fetchMenuRequest, string) {
	request := &fetchMenuRequest{
		Category: strings.TrimSpace(c.Query("category")),
		Search:   strings.TrimSpace(c.Query("search")),
		Limit:    defaultMenuLimit,
		Offset:   0,
	}

	isAvailableStr := strings.TrimSpace(c.Query("is_available"))
	if isAvailableStr != "" {
		isAvailable, err := strconv.ParseBool(isAvailableStr)
		if err != nil {
			return nil, "is_available must be a boolean"
		}
		request.IsAvailable = &isAvailable
	}

	pageSizeStr := strings.TrimSpace(c.Query("page_size"))
	limitStr := strings.TrimSpace(c.Query("limit"))
	if pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			return nil, "page_size must be a positive integer"
		}
		request.Limit = pageSize
	} else if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return nil, "limit must be a positive integer"
		}
		request.Limit = limit
	}

	if request.Limit > maxMenuLimit {
		request.Limit = maxMenuLimit
	}

	pageStr := strings.TrimSpace(c.Query("page"))
	offsetStr := strings.TrimSpace(c.Query("offset"))
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return nil, "page must be a positive integer"
		}
		request.Offset = (page - 1) * request.Limit
	} else if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return nil, "offset must be a non-negative integer"
		}
		request.Offset = offset
	}

	return request, ""
}

func (h *MenuHandler) Create(c *gin.Context) {
	var item domain.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if msg := validateMenuItem(&item); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if err := h.MenuUsecase.Create(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *MenuHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	item, err := h.MenuUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu item"})
		return
	}

	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *MenuHandler) Fetch(c *gin.Context) {
	request, validationErr := h.parseFetchMenuRequest(c)
	if validationErr != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}

	filter := domain.MenuFilter{
		Category:    request.Category,
		IsAvailable: request.IsAvailable,
		Search:      request.Search,
		Limit:       request.Limit,
		Offset:      request.Offset,
	}

	result, err := h.MenuUsecase.Fetch(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch menu items"})
		return
	}

	c.JSON(http.StatusOK, fetchMenuResponse{
		Data:   result.Items,
		Limit:  filter.Limit,
		Offset: filter.Offset,
		Total:  result.Total,
	})
}

func (h *MenuHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var item domain.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if msg := validateMenuItem(&item); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	item.ID = id
	if err := h.MenuUsecase.Update(c.Request.Context(), &item); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.MenuUsecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu item"})
		return
	}

	c.Status(http.StatusNoContent)
}
