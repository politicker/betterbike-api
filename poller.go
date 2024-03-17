package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/politicker/betterbike-api/internal/citibike"
	"github.com/politicker/betterbike-api/internal/db"
	"go.uber.org/zap"
)

//go:embed query.graphql
var citibikeAPIQuery string

const (
	baseURL = "https://citibikenyc.com/bikesharefe-gql"
)

type Poller struct {
	queries      *db.Queries
	pollDuration time.Duration
	logger       *zap.Logger
}

func NewPoller(queries *db.Queries, logger *zap.Logger, pollDuration time.Duration) Poller {
	if pollDuration < 1*time.Minute {
		pollDuration = 1 * time.Minute
	}

	return Poller{
		logger:       logger,
		queries:      queries,
		pollDuration: pollDuration,
	}
}

func (p *Poller) Start() {
	for {
		err := p.poll()
		if err != nil {
			p.logger.Error("poller error", zap.Error(err))
		}

		<-time.After(1 * time.Minute)
	}
}

func (p *Poller) poll() error {
	data, err := p.fetchStationData()
	if err != nil {
		return err
	}

	err = p.insertStationData(data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Poller) insertStationData(response *citibike.APIResponse) error {
	p.logger.Info(
		fmt.Sprintf("inserting station data for %d stations", len(response.Data.Supply.Stations)),
		zap.Int("stationCount", len(response.Data.Supply.Stations)),
	)

	for _, station := range response.Data.Supply.Stations {
		ebikesJson, err := json.Marshal(station.Ebikes)
		if err != nil {
			p.logger.Error("error marshalling ebikes", zap.Error(err))
			return err
		}

		err = p.queries.InsertStation(context.TODO(), db.InsertStationParams{
			ID:                 station.StationId,
			Name:               station.StationName,
			Lat:                station.Location.Lat,
			Lon:                station.Location.Lng,
			BikesAvailable:     int32(station.BikesAvailable),
			EbikesAvailable:    int32(station.EbikesAvailable),
			Ebikes:             ebikesJson,
			BikeDocksAvailable: int32(station.BikeDocksAvailable),
			LastUpdatedMs:      station.LastUpdatedMs,
			IsOffline:          station.IsOffline,
		})
		if err != nil {
			sentry.CaptureException(err)
			p.logger.Error("error inserting station", zap.Error(err))
			return err
		}

		err = p.queries.InsertStationTimeseries(context.TODO(), db.InsertStationTimeseriesParams{
			ID:                 station.StationId,
			Name:               station.StationName,
			Lat:                station.Location.Lat,
			Lon:                station.Location.Lng,
			BikesAvailable:     int32(station.BikesAvailable),
			EbikesAvailable:    int32(station.EbikesAvailable),
			Ebikes:             ebikesJson,
			BikeDocksAvailable: int32(station.BikeDocksAvailable),
			LastUpdatedMs:      station.LastUpdatedMs,
			IsOffline:          station.IsOffline,
		})
		if err != nil {
			sentry.CaptureException(err)
			p.logger.Error("error inserting station_timeseries", zap.Error(err))
			return err
		}
	}

	p.logger.Info("inserted station and timeseries data")

	return nil
}

type StationSupplyRequest struct {
	OperationName string    `json:"operationName"`
	Variables     Variables `json:"variables"`
	Query         string    `json:"query"`
}

type Variables struct {
	Input Input `json:"input"`
}

type Input struct {
	RegionCode        string `json:"regionCode"`
	RideablePageLimit int    `json:"rideablePageLimit"`
}

func (p *Poller) fetchStationData() (*citibike.APIResponse, error) {
	jsonPayload, err := json.Marshal(StationSupplyRequest{
		OperationName: "GetSupply",
		Variables: Variables{
			Input: Input{
				RegionCode:        "BKN",
				RideablePageLimit: 1000,
			},
		},
		Query: citibikeAPIQuery,
	})
	if err != nil {
		return nil, err
	}

	p.logger.Info("fetching station data")
	reader := bytes.NewReader(jsonPayload)
	resp, err := http.Post(baseURL, "application/json", reader)
	if err != nil {
		p.logger.Error("error fetching station data", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("citibike api returned status code %d", resp.StatusCode)
		p.logger.Error("citibike api error", zap.Error(err), zap.Int("statusCode", resp.StatusCode))
		return nil, err
	}

	result := citibike.APIResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		p.logger.Error("error decoding station data", zap.Error(err))
		return nil, err
	}

	p.logger.Info("fetched station data")
	return &result, nil
}
