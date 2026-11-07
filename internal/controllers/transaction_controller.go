package controllers

import (
	"net/http"

	"github.com/codekop/gomvc/internal/helper"
	"github.com/codekop/gomvc/internal/repositories"
	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	repository *repositories.TransactionRepository
}

// NewTransactionController creates the read-only transaction controller.
func NewTransactionController(repository *repositories.TransactionRepository) *TransactionController {
	return &TransactionController{repository: repository}
}

// Latest returns the newest simple transaction rows for the legacy endpoint.
func (c *TransactionController) Latest(ctx *gin.Context) {
	c.Index(ctx)
}

// Index lists simple transactions with complete pagination metadata.
func (c *TransactionController) Index(ctx *gin.Context) {
	page, perPage, err := helper.Paging(ctx.Query("page"), ctx.Query("per_page"), ctx.Query("limit"))
	if err != nil {
		helper.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	items, err := c.repository.ListPage(ctx.Request.Context(), perPage, helper.Offset(page, perPage))
	if err != nil {
		helper.InternalError(ctx, err, http.StatusServiceUnavailable, "Database query failed")
		return
	}
	total, err := c.repository.Count(ctx.Request.Context())
	if err != nil {
		helper.InternalError(ctx, err, http.StatusServiceUnavailable, "Database count failed")
		return
	}
	helper.SuccessPage(ctx, http.StatusOK, items, helper.MakePageInfo(page, perPage, total))
}
