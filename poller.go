package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apoliticker/citibike/citibike"
	"github.com/apoliticker/citibike/db"
	"github.com/apoliticker/citibike/logger"
)

const (
	baseURL          = "https://account.citibikenyc.com/bikesharefe-gql"
	citibikeAPIQuery = "query GetSystemSupply { supply { stations { stationId stationName location { lat lng __typename } bikesAvailable bikeDocksAvailable ebikesAvailable scootersAvailable totalBikesAvailable totalRideablesAvailable isValet isOffline isLightweight displayMessages siteId ebikes { batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } scooters { batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } lastUpdatedMs __typename } rideables { location { lat lng __typename } rideableType batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } notices { localizedTitle localizedDescription url __typename } requestErrors { localizedTitle localizedDescription url __typename } __typename }}"
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
			p.logger.Error(fmt.Sprintf("error polling citibike api: %s", err))
		}

		<-time.After(1 * time.Minute)
	}
}

func (p *Poller) poll() error {
	p.logger.Info("polling citibike api")

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
	p.logger.Info("inserting station data")

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

	reader := bytes.NewReader(jsonPayload)
	resp, err := http.Post(baseURL, "application/json", reader)
	if err != nil {
		return nil, err
	}

	result := citibike.APIResponse{}
	json.NewDecoder(resp.Body).Decode(&result)

	return &result, nil
}
