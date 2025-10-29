package httpdelivery

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/romanovevgeniy/weather-service/internal/usecase"
)

type Server struct {
	getLatest *usecase.GetLatestReading
}

func NewServer(getLatest *usecase.GetLatestReading) *Server {
	return &Server{getLatest: getLatest}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{city}", s.handleGetLatest)
	return r
}

func (s *Server) handleGetLatest(w http.ResponseWriter, r *http.Request) {
	city := chi.URLParam(r, "city")
	reading, err := s.getLatest.Execute(r.Context(), city)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reading)
}
