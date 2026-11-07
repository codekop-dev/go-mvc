package models

import "fmt"

type Item struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Total     float64 `json:"total"`
	CreatedAt any     `json:"created_at"`
}

// ItemFromMap converts a generic CRUD row into the public item model.
func ItemFromMap(row map[string]any) Item {
	return Item{
		ID:        integer(row["id"]),
		Name:      fmt.Sprint(row["name"]),
		Total:     number(row["total"]),
		CreatedAt: row["created_at"],
	}
}

// integer normalizes database driver numeric values into int64.
func integer(value any) int64 {
	switch value := value.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	default:
		var result int64
		_, _ = fmt.Sscan(fmt.Sprint(value), &result)
		return result
	}
}

// number normalizes database driver numeric values into float64.
func number(value any) float64 {
	switch value := value.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	default:
		var result float64
		_, _ = fmt.Sscan(fmt.Sprint(value), &result)
		return result
	}
}
