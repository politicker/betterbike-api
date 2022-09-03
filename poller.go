package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apoliticker/citibike/citibike"
	"github.com/apoliticker/citibike/db"
	"github.com/apoliticker/citibike/logger"
)

//go:embed query.graphql
var citibikeAPIQuery string

const (
	baseURL = "https://account.citibikenyc.com/bikesharefe-gql"
)

type Poller struct {
	queries      *db.Queries
	pollDuration time.Duration
	logger       logger.LogWriter
}

func NewPoller(queries *db.Queries, pollDuration time.Duration) Poller {
	if pollDuration < 1*time.Minute {
		pollDuration = 1 * time.Minute
	}

	return Poller{
		logger:       logger.New("poller"),
		queries:      queries,
		pollDuration: pollDuration,
	}
}

func (p *Poller) Start() {
	for {
		err := p.poll()
		if err != nil {
			p.logger.Error(fmt.Sprintf("poller: %s", err))
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
	p.logger.Info(fmt.Sprintf("inserting station data for %d stations", len(response.Data.Supply.Stations)))

	for _, station := range response.Data.Supply.Stations {
		ebikesJson, err := json.Marshal(station.Ebikes)
		if err != nil {
			p.logger.Error(fmt.Sprintf("error marshalling ebikes: %s", err))
			return err
		}

		err = p.queries.InsertStation(context.TODO(), db.InsertStationParams{
			ID:                 station.StationId,
			Name:               station.StationName,
			Lat:                station.Location.Lat,
			Lon:                station.Location.Lng,
			EbikesAvailable:    int32(station.EbikesAvailable),
			BikeDocksAvailable: int32(station.BikeDocksAvailable),
			Ebikes:             ebikesJson,
		})
		if err != nil {
			p.logger.Error(fmt.Sprintf("error inserting station: %s", err))
			return err
		}
	}

	p.logger.Info("inserted station data")
	return nil
}

func (p *Poller) fetchStationData() (*citibike.APIResponse, error) {
	jsonPayload, err := json.Marshal(map[string]string{
		"query": citibikeAPIQuery,
	})
	if err != nil {
		return nil, err
	}

	p.logger.Info("fetching station data")
	reader := bytes.NewReader(jsonPayload)
	resp, err := http.Post(baseURL, "application/json", reader)
	if err != nil {
		p.logger.Error("error fetching station data %v", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("citibike api returned status code %d", resp.StatusCode)
	}

	result := citibike.APIResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		p.logger.Error("error decoding station data %v", err)
		return nil, err
	}

	p.logger.Info("fetched station data")
	return &result, nil
}
