package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/apoliticker/citibike/citibike"
	"github.com/apoliticker/citibike/db"
)

const (
	baseURL          = "https://account.citibikenyc.com/bikesharefe-gql"
	citibikeAPIQuery = "query GetSystemSupply { supply { stations { stationId stationName location { lat lng __typename } bikesAvailable bikeDocksAvailable ebikesAvailable scootersAvailable totalBikesAvailable totalRideablesAvailable isValet isOffline isLightweight displayMessages siteId ebikes { batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } scooters { batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } lastUpdatedMs __typename } rideables { location { lat lng __typename } rideableType batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } notices { localizedTitle localizedDescription url __typename } requestErrors { localizedTitle localizedDescription url __typename } __typename }}"
)

type Poller struct {
	queries *db.Queries
}

func NewPoller(queries *db.Queries) Poller {
	return Poller{
		queries: queries,
	}
}

func (p *Poller) Start() {
	for {
		err := p.poll()
		if err != nil {
			log.Println("an error!")
			log.Println(err)
		}

		<-time.After(1 * time.Minute)
	}
}

func (p *Poller) poll() error {
	log.Println("requesting citibike api")

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
	for i, station := range response.Data.Supply.Stations {
		log.Printf("inserting station %d: %s", i, station.StationName)

		ebikesJson, err := json.Marshal(station.Ebikes)
		if err != nil {
			log.Printf("error marshalling ebikes: %s", err)
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
			log.Fatalf("error inserting station: %s", err)
			return err
		}

		log.Printf("inserted station %d: %s", i, station.StationName)
	}

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
