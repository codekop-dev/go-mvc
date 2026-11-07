package routes

import (
	"database/sql"
	"log/slog"

	"github.com/codekop/gomvc/internal/app/config"
	"github.com/codekop/gomvc/internal/controllers"
	"github.com/codekop/gomvc/internal/repositories"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Config config.Config
	DB     *sql.DB
	Logger *slog.Logger
}

// Register wires controllers into the module-specific route files.
func Register(router *gin.Engine, deps Dependencies) {
	apiController := controllers.NewAPIController(deps.Config, deps.DB)
	itemController := controllers.NewItemController(repositories.NewItemRepository(deps.DB, deps.Config.Database.Driver))
	transactionController := controllers.NewTransactionController(repositories.NewTransactionRepository(deps.DB))

	api := router.Group("/api")
	registerWebRoutes(router, apiController)
	registerAPIRoutes(api, apiController)
	registerItemsRoutes(api, itemController)
	registerTransaksiRoutes(api, transactionController)
}
