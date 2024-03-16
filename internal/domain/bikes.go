package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/getsentry/sentry-go"
	"github.com/politicker/betterbike-api/internal/api"
	"github.com/politicker/betterbike-api/internal/citibike"
	"github.com/politicker/betterbike-api/internal/db"
	"go.uber.org/zap"
)

type BikesRepo struct {
	logger  *zap.Logger
	queries *db.Queries
}

func NewBikesRepo(queries *db.Queries, logger *zap.Logger) *BikesRepo {
	return &BikesRepo{
		queries: queries,
		logger:  logger,
	}
}

const MetersToFeet = 3.28084

// GetNearbyStationEbikes returns a list of nearby
// ebike availability based on the users location.
// It returns a list of api.Home structs. It's used by the
// iOS app.
func (b *BikesRepo) GetNearbyStationEbikes(ctx context.Context, params db.GetStationsParams) ([]api.Station, error) {
	var viewStations []api.Station

	stations, err := b.queries.GetStations(ctx, params)
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	for _, station := range stations {
		var bikes []api.Bike
		var ebikes []citibike.Ebike

		err := json.Unmarshal(station.Ebikes, &ebikes)
		if err != nil {
			return nil, err
		}

		for idx, bike := range ebikes {
			quarter := int((float64(bike.BatteryStatus.Percent)/100)*4) * 25
			var isNextGen = bike.MaxDistance() > 25

			var color string
			switch {
			case bike.BatteryStatus.Percent <= 33:
				color = "var(--danger-color)"
			case bike.BatteryStatus.Percent <= 66:
				color = "var(--warning-color)"
			default:
				color = "var(--success-color)"
			}

			bikes = append(bikes, api.Bike{
				ID:                fmt.Sprintf("%s-%d", station.ID, idx),
				BatteryIcon:       fmt.Sprintf("battery.%d", quarter),
				BatteryColor:      template.CSS(color),
				BatteryPercentage: fmt.Sprintf("%d%%", bike.BatteryStatus.Percent),
				Range:             fmt.Sprintf("%d %s", bike.BatteryStatus.DistanceRemaining.Value, bike.BatteryStatus.DistanceRemaining.Unit),
				IsNextGen:         isNextGen,
			})
		}

		viewStations = append(viewStations, api.Station{
			ID:             station.ID,
			Name:           station.Name,
			BikeCount:      fmt.Sprint(station.EbikesAvailable),
			Bikes:          bikes,
			Lat:            station.Lat,
			Lon:            station.Lon,
			Distance:       station.Distance,
			PrettyDistance: fmt.Sprintf("%d feet", int(station.Distance*MetersToFeet)),
			CreatedAt:      station.CreatedAt,
		})
	}

	return viewStations, nil
}

// GetNearbyStations returns a list of nearby stations
// based on the users location. It returns a list of Station
// structs. It's used by the website.
func (b *BikesRepo) GetNearbyStations() {}
