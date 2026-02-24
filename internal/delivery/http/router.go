package http

import (
	"coffee-shop-pos/internal/delivery/http/handler"
	"coffee-shop-pos/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine, h *handler.MenuHandler) {
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.BodySizeLimit())

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
