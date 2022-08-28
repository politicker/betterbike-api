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
				BikesAvailable          int           `json:"bikesAvailable"`
				BikeDocksAvailable      int           `json:"bikeDocksAvailable"`
				EbikesAvailable         int           `json:"ebikesAvailable"`
				ScootersAvailable       int           `json:"scootersAvailable"`
				TotalBikesAvailable     int           `json:"totalBikesAvailable"`
				TotalRideablesAvailable int           `json:"totalRideablesAvailable"`
				IsValet                 bool          `json:"isValet"`
				IsOffline               bool          `json:"isOffline"`
				IsLightweight           bool          `json:"isLightweight"`
				DisplayMessages         []string      `json:"displayMessages"`
				SiteId                  string        `json:"siteId"`
				Ebikes                  []Ebike       `json:"ebikes"`
				Scooters                []interface{} `json:"scooters"`
				LastUpdatedMs           int64         `json:"lastUpdatedMs"`
			} `json:"stations"`
			Rideables     []interface{} `json:"rideables"`
			Notices       []interface{} `json:"notices"`
			RequestErrors []interface{} `json:"requestErrors"`
		} `json:"supply"`
	} `json:"data"`
}
