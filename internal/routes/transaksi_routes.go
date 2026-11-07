package routes

import (
	"github.com/codekop/gomvc/internal/controllers"
	"github.com/gin-gonic/gin"
)

// registerTransaksiRoutes registers the read-only transaction test endpoint.
func registerTransaksiRoutes(api *gin.RouterGroup, controller *controllers.TransactionController) {
	api.GET("/db/transaksi", controller.Latest)
}
