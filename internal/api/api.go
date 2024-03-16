package api

import (
	"html/template"
	"time"
)

type Bike struct {
	ID                string       `json:"id"`
	Range             string       `json:"range"`
	BatteryIcon       string       `json:"batteryIcon"`
	BatteryColor      template.CSS `json:"batteryColor"`
	BatteryPercentage string       `json:"batteryPercentage"`
	IsNextGen         bool         `json:"isNextGen"`
}

type Station struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BikeCount      string    `json:"bikeCount"`
	Bikes          []Bike    `json:"bikes"`
	Lat            float64   `json:"lat"`
	Lon            float64   `json:"lon"`
	Distance       float64   `json:"distance"`
	PrettyDistance string    `json:"prettyDistance"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Home struct {
	LastUpdated time.Time `json:"lastUpdated"`
	Stations    []Station `json:"stations"`
}
