package helper

import (
	"fmt"
	"strconv"
	"strings"
)

// PageInfo contains the common pagination values returned by list endpoints.
type PageInfo struct {
	CurrentPage int    `json:"current_page"`
	PerPage     int    `json:"per_page"`
	Total       int64  `json:"total"`
	LastPage    int    `json:"last_page"`
	From        *int64 `json:"from"`
	To          *int64 `json:"to"`
	NextPage    *int   `json:"next_page"`
	PrevPage    *int   `json:"prev_page"`
}

// Paging reads page/per_page and supports limit as a familiar alias.
func Paging(pageRaw, perPageRaw, limitRaw string) (int, int, error) {
	page := 1
	if strings.TrimSpace(pageRaw) != "" {
		parsed, err := strconv.Atoi(pageRaw)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}
	perPage := perPageRaw
	if strings.TrimSpace(perPage) == "" {
		perPage = limitRaw
	}
	parsedPerPage, err := Limit(perPage, 20, 100)
	if err != nil {
		return 0, 0, fmt.Errorf("per_page must be an integer")
	}
	return page, parsedPerPage, nil
}

// Offset returns the SQL offset for a one-based page number.
func Offset(page, perPage int) int {
	return (page - 1) * perPage
}

// MakePageInfo creates complete metadata for a paginated response.
func MakePageInfo(page, perPage int, total int64) PageInfo {
	lastPage := 1
	if total > 0 {
		lastPage = int((total + int64(perPage) - 1) / int64(perPage))
	}
	info := PageInfo{CurrentPage: page, PerPage: perPage, Total: total, LastPage: lastPage}
	if total > 0 && page <= lastPage {
		from := int64((page-1)*perPage + 1)
		to := from + int64(perPage) - 1
		if to > total {
			to = total
		}
		info.From = &from
		info.To = &to
	}
	if page < lastPage {
		next := page + 1
		info.NextPage = &next
	}
	if page > 1 && page <= lastPage {
		previous := page - 1
		info.PrevPage = &previous
	}
	return info
}

type Validator struct {
	errors map[string][]string
}

// NewValidator creates an empty field-error collector for a controller request.
func NewValidator() *Validator {
	return &Validator{errors: make(map[string][]string)}
}

// Add appends a validation message to a field.
func (v *Validator) Add(field, message string) {
	v.errors[field] = append(v.errors[field], message)
}

// RequiredString trims a pointer string and records an error when it is empty.
func (v *Validator) RequiredString(field string, value *string) string {
	trimmed, ok := RequiredString(value)
	if !ok {
		v.Add(field, "field is required")
	}
	return trimmed
}

// NonNegativeNumber requires a number that is zero or greater.
func (v *Validator) NonNegativeNumber(field string, value *float64) {
	if value == nil {
		v.Add(field, "field is required")
		return
	}
	if *value < 0 {
		v.Add(field, "must be greater than or equal to 0")
	}
}

// HasErrors reports whether any validation rule failed.
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns validation messages grouped by input field.
func (v *Validator) Errors() map[string][]string {
	return v.errors
}

// PositiveID parses a route ID and accepts only positive integers.
func PositiveID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}

// Limit parses a query limit and clamps it to the configured maximum.
func Limit(raw string, fallback, maximum int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("limit must be an integer")
	}
	if limit < 1 {
		return 1, nil
	}
	if limit > maximum {
		return maximum, nil
	}
	return limit, nil
}

// RequiredString returns a trimmed value and whether it is non-empty.
func RequiredString(value *string) (string, bool) {
	if value == nil {
		return "", false
	}
	trimmed := strings.TrimSpace(*value)
	return trimmed, trimmed != ""
}

// NonNegative reports whether a numeric pointer exists and is not negative.
func NonNegative(value *float64) bool {
	return value != nil && *value >= 0
}
