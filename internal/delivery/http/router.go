package http

import (
	"coffee-shop-pos/internal/delivery/http/handler"
	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine, h *handler.MenuHandler) {
	api := r.Group("/api/v1")
	{
		menu := api.Group("/menu")
		{
			menu.POST("", h.Create)
			menu.GET("", h.Fetch)
			menu.GET("/:id", h.GetByID)
			menu.PUT("/:id", h.Update)
			menu.DELETE("/:id", h.Delete)
		}
	}
}
