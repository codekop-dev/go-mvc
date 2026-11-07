package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/codekop/gomvc/internal/app/config"
	"github.com/codekop/gomvc/internal/middleware"
	"github.com/codekop/gomvc/internal/routes"
	"github.com/gin-gonic/gin"
)

type App struct {
	config config.Config
	db     *sql.DB
	logger *slog.Logger
}

// New wires application dependencies without opening another database connection.
func New(cfg config.Config, db *sql.DB, logger *slog.Logger) *App {
	return &App{config: cfg, db: db, logger: logger}
}

// Handler builds the Gin engine, attaches middleware, and registers all routes.
func (a *App) Handler() http.Handler {
	if a.config.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.ErrorLogger(a.logger), middleware.RequestLogger(a.logger), middleware.SecurityHeaders(a.config.IsDevelopment()))
	routes.Register(router, routes.Dependencies{
		Config: a.config,
		DB:     a.db,
		Logger: a.logger,
	})
	return router
}
