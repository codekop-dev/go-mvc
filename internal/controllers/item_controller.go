package controllers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/codekop/gomvc/internal/helper"
	"github.com/codekop/gomvc/internal/repositories"
	"github.com/gin-gonic/gin"
)

type ItemController struct {
	repository *repositories.ItemRepository
}

type itemInput struct {
	Name  *string  `json:"name"`
	Total *float64 `json:"total"`
}

// NewItemController creates a CRUD controller with its repository dependency.
func NewItemController(repository *repositories.ItemRepository) *ItemController {
	return &ItemController{repository: repository}
}

// Index lists sample items with a bounded limit query parameter.
func (c *ItemController) Index(ctx *gin.Context) {
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

// Show returns one sample item by route id.
func (c *ItemController) Show(ctx *gin.Context) {
	id, ok := controllerID(ctx)
	if !ok {
		return
	}
	item, err := c.repository.Find(ctx.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		helper.Error(ctx, http.StatusNotFound, "Item not found")
		return
	}
	if err != nil {
		helper.InternalError(ctx, err, http.StatusServiceUnavailable, "Database query failed")
		return
	}
	helper.Success(ctx, http.StatusOK, item)
}

// Store validates and inserts one sample item.
func (c *ItemController) Store(ctx *gin.Context) {
	var body itemInput
	if !helper.DecodeJSON(ctx, &body) {
		return
	}
	validator := helper.NewValidator()
	name := validator.RequiredString("name", body.Name)
	validator.NonNegativeNumber("total", body.Total)
	if validator.HasErrors() {
		helper.ValidationError(ctx, validator.Errors())
		return
	}
	id, err := c.repository.Create(ctx.Request.Context(), name, *body.Total)
	if err != nil {
		helper.InternalError(ctx, err, http.StatusServiceUnavailable, "Database insert failed")
		return
	}
	helper.Success(ctx, http.StatusCreated, map[string]any{"id": id})
}

// Update applies a partial update to a sample item.
func (c *ItemController) Update(ctx *gin.Context) {
	id, ok := controllerID(ctx)
	if !ok {
		return
	}
	var body itemInput
	if !helper.DecodeJSON(ctx, &body) {
		return
	}
	validator := helper.NewValidator()
	values := make(map[string]any)
	if body.Name != nil {
		name := validator.RequiredString("name", body.Name)
		values["name"] = name
	}
	if body.Total != nil {
		validator.NonNegativeNumber("total", body.Total)
		values["total"] = *body.Total
	}
	if len(values) == 0 {
		validator.Add("body", "at least one field is required")
	}
	if validator.HasErrors() {
		helper.ValidationError(ctx, validator.Errors())
		return
	}
	updated, err := c.repository.Update(ctx.Request.Context(), id, values)
	if err != nil {
		helper.InternalError(ctx, err, http.StatusServiceUnavailable, "Database update failed")
		return
	}
	if !updated {
		helper.Error(ctx, http.StatusNotFound, "Item not found")
		return
	}
	helper.Success(ctx, http.StatusOK, map[string]any{"id": id})
}

// Destroy deletes one sample item by route id.
func (c *ItemController) Destroy(ctx *gin.Context) {
	id, ok := controllerID(ctx)
	if !ok {
		return
	}
	deleted, err := c.repository.Delete(ctx.Request.Context(), id)
	if err != nil {
		helper.InternalError(ctx, err, http.StatusServiceUnavailable, "Database delete failed")
		return
	}
	if !deleted {
		helper.Error(ctx, http.StatusNotFound, "Item not found")
		return
	}
	ctx.Status(http.StatusNoContent)
}

// controllerID parses and validates the common CRUD route parameter.
func controllerID(ctx *gin.Context) (int64, bool) {
	id, err := helper.PositiveID(ctx.Param("id"))
	if err != nil {
		helper.Error(ctx, http.StatusBadRequest, err.Error())
		return 0, false
	}
	return id, true
}
