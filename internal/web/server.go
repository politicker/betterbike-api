package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/politicker/betterbike-api/internal/api"
	"github.com/politicker/betterbike-api/internal/db"
	"github.com/politicker/betterbike-api/internal/domain"
	"go.uber.org/zap"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templates embed.FS

type Server struct {
	logger    *zap.Logger
	queries   *db.Queries
	port      string
	bikesRepo *domain.BikesRepo
}

func NewServer(ctx context.Context, logger *zap.Logger, queries *db.Queries, port string) Server {
	return Server{
		queries:   queries,
		port:      port,
		logger:    logger,
		bikesRepo: domain.NewBikesRepo(queries, logger),
	}
}

func (s *Server) Start() error {
	http.Handle("GET /static/", http.FileServer(http.FS(staticFiles)))

	http.HandleFunc("GET /api", s.GetBikes)
	http.HandleFunc("GET /", s.indexHandler)
	http.HandleFunc("GET /bikes", s.bikesHandler)

	s.logger.Info("listening", zap.String("port", s.port))
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) renderError(w http.ResponseWriter, message string, errorCode string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.logger.Error(message)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":     message,
		"errorCode": errorCode,
	})
}

func (s *Server) GetBikes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	start := time.Now()

	w.Header().Set("Content-Type", "application/json")

	var stationParams db.GetStationsParams
	err := json.NewDecoder(r.Body).Decode(&stationParams)
	if err != nil {
		s.renderError(w, "lat and lon are required", "invalid-json-payload")
		return
	}

	if stationParams.Lat == 0 || stationParams.Lon == 0 {
		s.renderError(w, fmt.Sprintf("invalid lat or lon: %f, %f", stationParams.Lat, stationParams.Lon), "missing-coords")
		return
	}

	stations, err := s.bikesRepo.GetNearbyStationEbikes(ctx, stationParams)
	if err != nil {
		s.renderError(w, "error fetching stations", "internal-error")
		return
	}
	if len(stations) == 0 {
		s.renderError(w, "No ebikes nearby. Are you in New York City?", "too-far-away")
		return
	}

	err = json.NewEncoder(w).Encode(api.Home{
		LastUpdated: stations[0].CreatedAt,
		Stations:    stations,
	})
	if err != nil {
		s.renderError(w, "error encoding response", "internal-error")
		return
	}

	s.logger.Info(
		fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, time.Since(start)),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", time.Since(start)),
	)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl, err := template.ParseFiles(
		"index.html",
		"layout.html",
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) bikesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("bikes.html", "layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var stationParams db.GetStationsParams
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		s.renderError(w, "invalid lat", "invalid-lat")
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		s.renderError(w, "invalid lon", "invalid-lon")
		return
	}

	stationParams.Lat = lat
	stationParams.Lon = lon

	if stationParams.Lat == 0 || stationParams.Lon == 0 {
		s.renderError(w, fmt.Sprintf("invalid lat or lon: %f, %f", stationParams.Lat, stationParams.Lon), "missing-coords")
		return
	}

	stations, err := s.bikesRepo.GetNearbyStationEbikes(r.Context(), stationParams)
	if err != nil {
		s.renderError(w, "error fetching stations", "internal-error")
		return
	}
	if len(stations) == 0 {
		s.renderError(w, "No ebikes nearby. Are you in New York City?", "too-far-away")
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", api.Home{
		LastUpdated: stations[0].CreatedAt,
		Stations:    stations,
		Lat:         &lat,
		Lon:         &lon})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
