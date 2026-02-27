package http

import (
	"coffee-shop-pos/internal/delivery/http/handler"
	"coffee-shop-pos/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(r *gin.Engine, menuHandler *handler.MenuHandler, orderHandler *handler.OrderHandler) {
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.BodySizeLimit())

	api := r.Group("/api/v1")
	{
		menu := api.Group("/menu")
		{
			menu.POST("", menuHandler.Create)
			menu.GET("", menuHandler.Fetch)
			menu.GET("/:id", menuHandler.GetByID)
			menu.PUT("/:id", menuHandler.Update)
			menu.DELETE("/:id", menuHandler.Delete)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.Create)
			orders.GET("", orderHandler.List)
			orders.GET("/:id", orderHandler.GetByID)
			orders.PATCH("/:id/status", orderHandler.UpdateStatus)
		}
	}
}
