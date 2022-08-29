package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apoliticker/citibike/api"
	"github.com/apoliticker/citibike/citibike"
	"github.com/apoliticker/citibike/db"
	"github.com/apoliticker/citibike/logger"
)

type Server struct {
	logger  logger.LogWriter
	queries *db.Queries
	port    string
}

func NewServer(port string, queries *db.Queries) Server {
	return Server{
		logger:  logger.New("server"),
		queries: queries,
		port:    port,
	}
}

func (s *Server) Start() {
	http.Handle("/", WithLogging(s.GetBikes()))

	s.logger.Info(fmt.Sprintf("listening on: %s", s.port))
	http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) renderError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.logger.Error(fmt.Sprintf("client error: message=%s", message))

	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (s *Server) GetBikes() http.Handler {
	ctx := context.Background()

	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var stationParams db.GetStationsParams
		err := json.NewDecoder(r.Body).Decode(&stationParams)
		if err != nil {
			s.renderError(w, "lat and lon are required")
			return
		}

		if stationParams.Lat == 0 || stationParams.Lon == 0 {
			s.renderError(w, fmt.Sprintf("invalid lat or lon: %f, %f", stationParams.Lat, stationParams.Lon))
			return
		}

		stations, err := s.queries.GetStations(ctx, stationParams)
		if err != nil {
			s.renderError(w, err.Error())
			return
		}

		var viewStations []api.Station

		for _, station := range stations {
			var bikes []api.Bike
			var ebikes []citibike.Ebike

			err = json.Unmarshal(station.Ebikes, &ebikes)

			if err != nil {
				s.renderError(w, err.Error())
				return
			}

			for idx, bike := range ebikes {
				quarter := int((float64(bike.BatteryStatus.Percent)/100)*4) * 25

				bikes = append(bikes, api.Bike{
					ID:          fmt.Sprintf("%s-%d", station.ID, idx),
					BatteryIcon: fmt.Sprintf("battery.%d", quarter),
					Range:       fmt.Sprintf("%d %s", bike.BatteryStatus.DistanceRemaining.Value, bike.BatteryStatus.DistanceRemaining.Unit),
				})
			}

			viewStations = append(viewStations, api.Station{
				ID:        station.ID,
				Name:      station.Name,
				BikeCount: fmt.Sprint(station.EbikesAvailable),
				Bikes:     bikes,
			})
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.Home{
			LastUpdated: stations[0].CreatedAt,
			Stations:    viewStations,
		})
	}

	return http.HandlerFunc(fn)
}
