package view

import "time"

type Bike struct {
	Range       string `json:"range"`
	BatteryIcon string `json:"batteryIcon"`
}

type Station struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BikeCount string `json:"bikeCount"`
	Bikes     []Bike `json:"bikes"`
}

type Home struct {
	LastUpdated time.Time `json:"lastUpdated"`
	Stations    []Station `json:"stations"`
}