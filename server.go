package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"

	"github.com/apoliticker/citibike/api"
	"github.com/apoliticker/citibike/citibike"
	"github.com/apoliticker/citibike/db"
)

type Server struct {
	logger  *zap.Logger
	queries *db.Queries
	port    string
}

func NewServer(port string, queries *db.Queries, logger *zap.Logger) Server {
	return Server{
		queries: queries,
		port:    port,
		logger:  logger,
	}
}

func (s *Server) Start() {
	http.HandleFunc("/", s.GetBikes)

	s.logger.Info("listening", zap.String("port", s.port))
	http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) renderError(w http.ResponseWriter, message string, errorCode string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.logger.Error(message)

	json.NewEncoder(w).Encode(map[string]string{
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

	stations, err := s.queries.GetStations(ctx, stationParams)
	if err != nil {
		sentry.CaptureException(err)
		s.renderError(w, err.Error(), "get-stations-failed")
		return
	}

	var viewStations []api.Station

	for _, station := range stations {
		var bikes []api.Bike
		var ebikes []citibike.Ebike

		err = json.Unmarshal(station.Ebikes, &ebikes)
		if err != nil {
			s.renderError(w, err.Error(), "invalid-json-ebikes")
			return
		}

		for idx, bike := range ebikes {
			quarter := int((float64(bike.BatteryStatus.Percent)/100)*4) * 25
			var isNextGen = bike.MaxDistance() > 25

			bikes = append(bikes, api.Bike{
				ID:          fmt.Sprintf("%s-%d", station.ID, idx),
				BatteryIcon: fmt.Sprintf("battery.%d", quarter),
				Range:       fmt.Sprintf("%d %s", bike.BatteryStatus.DistanceRemaining.Value, bike.BatteryStatus.DistanceRemaining.Unit),
				IsNextGen:   isNextGen,
			})
		}

		viewStations = append(viewStations, api.Station{
			ID:        station.ID,
			Name:      station.Name,
			BikeCount: fmt.Sprint(station.EbikesAvailable),
			Bikes:     bikes,
			Lat:       station.Lat,
			Lon:       station.Lon,
			Distance:  station.Distance,
		})
	}

	if len(stations) == 0 {
		s.renderError(w, "No ebikes nearby. Are you in New York City?", "too-far-away")
		return
	}

	json.NewEncoder(w).Encode(api.Home{
		LastUpdated: stations[0].CreatedAt,
		Stations:    viewStations,
	})

	s.logger.Info(
		fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, time.Since(start)),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", time.Since(start)),
	)
}
