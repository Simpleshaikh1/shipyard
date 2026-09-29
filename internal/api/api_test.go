package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Simpleshaikh1/shipyard/internal/jobs"
)

func TestRoutes(t *testing.T) {
	srv := New(jobs.NewMemoryStore(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	h := srv.Routes()

	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"health", http.MethodGet, "/healthz", "", http.StatusOK},
		{"create ok", http.MethodPost, "/jobs", `{"type":"email","payload":"hi"}`, http.StatusCreated},
		{"create missing type", http.MethodPost, "/jobs", `{"payload":"hi"}`, http.StatusBadRequest},
		{"create bad json", http.MethodPost, "/jobs", `{`, http.StatusBadRequest},
		{"list", http.MethodGet, "/jobs", "", http.StatusOK},
		{"get missing", http.MethodGet, "/jobs/nope", "", http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("got %d, want %d (body: %s)", rec.Code, tc.want, rec.Body)
			}
		})
	}
}
