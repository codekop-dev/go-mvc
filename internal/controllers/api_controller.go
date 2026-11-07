package controllers

import (
	"database/sql"
	"net/http"
	"runtime"

	"github.com/codekop/gomvc/internal/app/config"
	"github.com/codekop/gomvc/internal/helper"
	"github.com/gin-gonic/gin"
)

type APIController struct {
	config config.Config
	db     *sql.DB
}

// NewAPIController creates the controller for stateless API examples and health checks.
func NewAPIController(cfg config.Config, db *sql.DB) *APIController {
	return &APIController{config: cfg, db: db}
}

// Welcome returns a small response proving that the application is running.
func (c *APIController) Welcome(ctx *gin.Context) {
	helper.Success(ctx, http.StatusOK, map[string]any{
		"name":    c.config.Name,
		"message": "Go MVC API is running",
		"docs":    "/api/health",
	})
}

// Health reports application metadata and the optional database status.
func (c *APIController) Health(ctx *gin.Context) {
	databaseStatus := "not_configured"
	status := http.StatusOK
	if c.db != nil {
		databaseStatus = "connected"
		if err := c.db.PingContext(ctx.Request.Context()); err != nil {
			_ = ctx.Error(err)
			databaseStatus = "unavailable"
			status = http.StatusServiceUnavailable
		}
	}
	helper.JSON(ctx, status, map[string]any{
		"code":    status,
		"success": status == http.StatusOK,
		"data": map[string]any{
			"name":     c.config.Name,
			"version":  c.config.Version,
			"go":       runtime.Version(),
			"env":      c.config.AppEnv,
			"timezone": c.config.Location().String(),
			"database": databaseStatus,
		},
	})
}

// Echo decodes JSON and sends the same payload back to the client.
func (c *APIController) Echo(ctx *gin.Context) {
	var payload map[string]any
	if !helper.DecodeJSON(ctx, &payload) {
		return
	}
	helper.Success(ctx, http.StatusOK, payload)
}

// Route demonstrates Gin path parameters and query-string access.
func (c *APIController) Route(ctx *gin.Context) {
	helper.Success(ctx, http.StatusOK, map[string]any{
		"parameters": map[string]string{"id": ctx.Param("id")},
		"query":      ctx.Request.URL.Query(),
	})
}
