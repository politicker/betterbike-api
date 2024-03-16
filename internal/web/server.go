package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/politicker/betterbike-api/internal/api"
	"github.com/politicker/betterbike-api/internal/db"
	"github.com/politicker/betterbike-api/internal/domain"
	"go.uber.org/zap"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templates embed.FS

type web struct {
	port      string
	title     string
	logger    *zap.Logger
	queries   *db.Queries
	bikesRepo *domain.BikesRepo
}

func NewWeb(ctx context.Context, logger *zap.Logger, queries *db.Queries, port string) *web {
	return &web{
		port:      port,
		logger:    logger,
		queries:   queries,
		bikesRepo: domain.NewBikesRepo(queries, logger),
	}
}

func (s *web) Start() error {
	http.HandleFunc("/", s.indexHandler)

	fs := http.FileServer(http.FS(staticFiles))
	http.Handle("/static/", fs)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *web) indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templates, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var stationParams db.GetStationsParams
	err = json.NewDecoder(r.Body).Decode(&stationParams)
	if err != nil {
		s.renderError(w, "lat and lon are required", "invalid-json-payload")
		return
	}

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

	err = tmpl.Execute(w, api.Home{LastUpdated: stations[0].CreatedAt, Stations: stations})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *web) renderError(w http.ResponseWriter, message string, errorCode string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.logger.Error(message)

	json.NewEncoder(w).Encode(map[string]string{
		"error":     message,
		"errorCode": errorCode,
	})
}
