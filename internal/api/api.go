package api

import (
	"html/template"
	"time"
)

type Ebike struct {
	ID                string       `json:"id"`
	Range             string       `json:"range"`
	BatteryIcon       string       `json:"batteryIcon"`
	BatteryColor      template.CSS `json:"batteryColor"`
	BatteryPercentage string       `json:"batteryPercentage"`
	IsNextGen         bool         `json:"isNextGen"`
}

type Station struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	EbikesAvailable    string    `json:"bikeCount"`
	BikesAvailable     int32     `json:"bikesAvailable"`
	BikeDocksAvailable int32     `json:"bikeDocksAvailable"`
	Ebikes             []Ebike   `json:"bikes"`
	Lat                float64   `json:"lat"`
	Lon                float64   `json:"lon"`
	Distance           float64   `json:"distance"`
	PrettyDistance     string    `json:"prettyDistance"`
	LastUpdated        time.Time `json:"lastUpdated"`
	PrettyLastUpdated  string    `json:"prettyLastUpdated"`
	CreatedAt          time.Time `json:"createdAt"`
}

type Home struct {
	Lat         *float64  `json:"lat"`
	Lon         *float64  `json:"lon"`
	LastUpdated time.Time `json:"lastUpdated"`
	Stations    []Station `json:"stations"`
}
