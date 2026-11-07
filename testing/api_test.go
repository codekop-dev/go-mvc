package integration_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codekop/gomvc/internal/app"
	"github.com/codekop/gomvc/internal/app/config"
)

// TestIntegrationAPI verifies the public HTTP contract without requiring a live database.
func TestIntegrationAPI(t *testing.T) {
	cfg := config.Load()
	cfg.AppEnv = "testing"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := app.New(cfg, nil, logger).Handler()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "health", method: http.MethodGet, path: "/api/health", wantStatus: http.StatusOK, wantBody: `"success":true`},
		{name: "route", method: http.MethodGet, path: "/api/route/demo-1?from=php", wantStatus: http.StatusOK, wantBody: `"id":"demo-1"`},
		{name: "echo", method: http.MethodPost, path: "/api/echo", body: `{"name":"Budi"}`, wantStatus: http.StatusOK, wantBody: `"name":"Budi"`},
		{name: "CRUD validation", method: http.MethodPost, path: "/api/items", body: `{"total":-1}`, wantStatus: http.StatusUnprocessableEntity, wantBody: `"name"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want substring %q", response.Body.String(), tt.wantBody)
			}
		})
	}
}
