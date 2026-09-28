// Package api wires HTTP routes to the jobs.Store.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Simpleshaikh1/shipyard/internal/jobs"
)

type Server struct {
	store jobs.Store
	log   *slog.Logger
}

func New(store jobs.Store, log *slog.Logger) *Server {
	return &Server{store: store, log: log}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)

	r.Get("/healthz", s.healthz)
	r.Route("/jobs", func(r chi.Router) {
		r.Post("/", s.createJob)
		r.Get("/", s.listJobs)
		r.Get("/{id}", s.getJob)
	})
	return r
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	job, err := s.store.Create(r.Context(), jobs.Job{Type: req.Type, Payload: req.Payload})
	if err != nil {
		s.log.Error("create job", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	s.log.Info("job created", "id", job.ID, "type", job.Type)
	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	job, err := s.store.Get(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, jobs.ErrNotFound) {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	if err != nil {
		s.log.Error("get job", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.List(r.Context())
	if err != nil {
		s.log.Error("list jobs", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
