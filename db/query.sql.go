// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: query.sql

package db

import (
	"context"
	"encoding/json"
)

const getStations = `-- name: GetStations :many
select
	id,
	name,
	lat,
	lon,
	ebikes_available,
	bike_docks_available,
	ebikes,
	ST_MakePoint(lat, lon) <-> ST_MakePoint( $1, $2 ) AS distance
from stations
order by distance asc
limit 10
`

type GetStationsParams struct {
	Lat interface{}
	Lon interface{}
}

type GetStationsRow struct {
	ID                 string
	Name               string
	Lat                float64
	Lon                float64
	EbikesAvailable    int32
	BikeDocksAvailable int32
	Ebikes             json.RawMessage
	Distance           interface{}
}

func (q *Queries) GetStations(ctx context.Context, arg GetStationsParams) ([]GetStationsRow, error) {
	rows, err := q.db.QueryContext(ctx, getStations, arg.Lat, arg.Lon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStationsRow
	for rows.Next() {
		var i GetStationsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Lat,
			&i.Lon,
			&i.EbikesAvailable,
			&i.BikeDocksAvailable,
			&i.Ebikes,
			&i.Distance,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertStation = `-- name: InsertStation :exec
insert into
	stations (
		id,
		name,
		lat,
		lon,
		ebikes_available,
		bike_docks_available,
		ebikes
	)
values (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7
)
ON CONFLICT (id) DO UPDATE
	SET
		name = EXCLUDED.name,
		lat = EXCLUDED.lat,
		lon = EXCLUDED.lon,
		ebikes_available = EXCLUDED.ebikes_available,
		bike_docks_available = EXCLUDED.bike_docks_available,
		ebikes = EXCLUDED.ebikes
`

type InsertStationParams struct {
	ID                 string
	Name               string
	Lat                float64
	Lon                float64
	EbikesAvailable    int32
	BikeDocksAvailable int32
	Ebikes             json.RawMessage
}

// InsertStation inserts citibike station data into the database.
func (q *Queries) InsertStation(ctx context.Context, arg InsertStationParams) error {
	_, err := q.db.ExecContext(ctx, insertStation,
		arg.ID,
		arg.Name,
		arg.Lat,
		arg.Lon,
		arg.EbikesAvailable,
		arg.BikeDocksAvailable,
		arg.Ebikes,
	)
	return err
}
