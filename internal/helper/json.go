package helper

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code       int    `json:"code"`
	Success    bool   `json:"success"`
	Data       any    `json:"data,omitempty"`
	Message    string `json:"message,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

// JSON writes any value with the standard JSON content type and HTTP status.
func JSON(c *gin.Context, status int, value any) {
	c.JSON(status, value)
}

// Success writes the project envelope used for successful API responses.
func Success(c *gin.Context, status int, data any) {
	JSON(c, status, Envelope{Code: status, Success: true, Data: data})
}

// SuccessPage writes a successful response with pagination metadata.
func SuccessPage(c *gin.Context, status int, data, pagination any) {
	JSON(c, status, Envelope{Code: status, Success: true, Data: data, Pagination: pagination})
}

// Error writes a compact project envelope for non-validation errors.
func Error(c *gin.Context, status int, message string) {
	JSON(c, status, Envelope{Code: status, Success: false, Message: message})
}

// InternalError records a private error in Gin context and returns a safe message.
func InternalError(c *gin.Context, err error, status int, message string) {
	if err != nil {
		_ = c.Error(err)
	}
	Error(c, status, message)
}

// ValidationError writes a PHP-friendly 422 response with field error arrays.
func ValidationError(c *gin.Context, errors map[string][]string) {
	JSON(c, 422, map[string]any{
		"code":    422,
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}

// DecodeJSON reads one bounded JSON body and rejects unknown or trailing data.
func DecodeJSON(c *gin.Context, destination any) bool {
	if c.Request.Body == nil {
		Error(c, 400, "Request body is required")
		return false
	}
	decoder := json.NewDecoder(io.LimitReader(c.Request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		Error(c, 400, "Invalid JSON body")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		Error(c, 400, "Request body must contain one JSON object")
		return false
	}
	return true
}
