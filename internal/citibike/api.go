package citibike

type Ebike struct {
	BatteryStatus struct {
		DistanceRemaining struct {
			Value int    `json:"value"`
			Unit  string `json:"unit"`
		} `json:"distanceRemaining"`
		Percent int `json:"percent"`
	} `json:"batteryStatus"`
}

type APIResponse struct {
	Data struct {
		Supply struct {
			Stations []struct {
				StationId   string `json:"stationId"`
				StationName string `json:"stationName"`
				Location    struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
				BikesAvailable      int     `json:"bikesAvailable"`
				BikeDocksAvailable  int     `json:"bikeDocksAvailable"`
				EbikesAvailable     int     `json:"ebikesAvailable"`
				TotalBikesAvailable int     `json:"totalBikesAvailable"`
				IsOffline           bool    `json:"isOffline"`
				Ebikes              []Ebike `json:"ebikes"`
				LastUpdatedMs       int64   `json:"lastUpdatedMs"`
			} `json:"stations"`
		} `json:"supply"`
	} `json:"data"`
}
