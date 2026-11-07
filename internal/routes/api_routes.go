package routes

import (
	"github.com/codekop/gomvc/internal/controllers"
	"github.com/gin-gonic/gin"
)

// registerWebRoutes registers non-prefixed web routes.
func registerWebRoutes(router *gin.Engine, controller *controllers.APIController) {
	router.GET("/", controller.Welcome)
}

// registerAPIRoutes registers stateless API demonstration endpoints.
func registerAPIRoutes(api *gin.RouterGroup, controller *controllers.APIController) {
	api.GET("/health", controller.Health)
	api.POST("/echo", controller.Echo)
	api.GET("/route/:id", controller.Route)
}
