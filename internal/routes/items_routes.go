package routes

import (
	"github.com/codekop/gomvc/internal/controllers"
	"github.com/gin-gonic/gin"
)

// registerItemsRoutes registers the sample CRUD resource endpoints.
func registerItemsRoutes(api *gin.RouterGroup, controller *controllers.ItemController) {
	items := api.Group("/items")
	{
		items.GET("", controller.Index)
		items.GET("/:id", controller.Show)
		items.POST("", controller.Store)
		items.PUT("/:id", controller.Update)
		items.PATCH("/:id", controller.Update)
		items.DELETE("/:id", controller.Destroy)
	}
}
