package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/apoliticker/citibike/db"
	_ "github.com/lib/pq"
)

type CitibikeAPIResponse struct {
	Data struct {
		Supply struct {
			Stations []struct {
				StationId   string `json:"stationId"`
				StationName string `json:"stationName"`
				Location    struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
				BikesAvailable          int      `json:"bikesAvailable"`
				BikeDocksAvailable      int      `json:"bikeDocksAvailable"`
				EbikesAvailable         int      `json:"ebikesAvailable"`
				ScootersAvailable       int      `json:"scootersAvailable"`
				TotalBikesAvailable     int      `json:"totalBikesAvailable"`
				TotalRideablesAvailable int      `json:"totalRideablesAvailable"`
				IsValet                 bool     `json:"isValet"`
				IsOffline               bool     `json:"isOffline"`
				IsLightweight           bool     `json:"isLightweight"`
				DisplayMessages         []string `json:"displayMessages"`
				SiteId                  string   `json:"siteId"`
				Ebikes                  []struct {
					BatteryStatus struct {
						DistanceRemaining struct {
							Value int    `json:"value"`
							Unit  string `json:"unit"`
						} `json:"distanceRemaining"`
						Percent int `json:"percent"`
					} `json:"batteryStatus"`
				} `json:"ebikes"`
				Scooters      []interface{} `json:"scooters"`
				LastUpdatedMs int64         `json:"lastUpdatedMs"`
			} `json:"stations"`
			Rideables     []interface{} `json:"rideables"`
			Notices       []interface{} `json:"notices"`
			RequestErrors []interface{} `json:"requestErrors"`
		} `json:"supply"`
	} `json:"data"`
}

var connectionString string
var database *sql.DB

func init() {
	connectionString = os.Getenv("DATABASE_CONN_STRING")

	var err error
	database, err = sql.Open("postgres", connectionString)
	if err != nil {
		panic("failed to connect to database")
	}
}

func poll() error {
	ctx := context.Background()
	log.Println("requesting citibike api")

	// marshal json
	jsonPayload, err := json.Marshal(map[string]string{
		"query": "query GetSystemSupply { supply { stations { stationId stationName location { lat lng __typename } bikesAvailable bikeDocksAvailable ebikesAvailable scootersAvailable totalBikesAvailable totalRideablesAvailable isValet isOffline isLightweight displayMessages siteId ebikes { batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } scooters { batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } lastUpdatedMs __typename } rideables { location { lat lng __typename } rideableType batteryStatus { distanceRemaining { value unit __typename } percent __typename } __typename } notices { localizedTitle localizedDescription url __typename } requestErrors { localizedTitle localizedDescription url __typename } __typename }}",
	})
	if err != nil {
		return err
	}

	// create io.reader from byte[]
	reader := bytes.NewReader(jsonPayload)

	// send a post request to the server
	resp, err := http.Post("https://account.citibikenyc.com/bikesharefe-gql", "application/json", reader)
	if err != nil {
		return err
	}

	result := CitibikeAPIResponse{}

	// unmarshal the response
	json.NewDecoder(resp.Body).Decode(&result)
	queries := db.New(database)

	for i, station := range result.Data.Supply.Stations {
		log.Printf("inserting station %d: %s", i, station.StationName)

		ebikesJson, err := json.Marshal(station.Ebikes)
		if err != nil {
			log.Printf("error marshalling ebikes: %s", err)
			return err
		}

		err = queries.InsertStation(ctx, db.InsertStationParams{
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

func startPoller() {
	go func() {
		for {
			err := poll()
			if err != nil {
				log.Println("an error!")
				log.Println(err)
			}

			// wait 1 minute
			<-time.After(1 * time.Minute)
		}
	}()
}

func main() {
	startPoller()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
