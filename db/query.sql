-- InsertStation inserts citibike station data into the database.
-- name: InsertStation :exec
insert into
	stations (
		id,
		name,
		lat,
		lon,
		ebikes_available,
		bike_docks_available,
		ebikes,
		created_at
	)
values (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7,
	now() at time zone 'utc'
)
ON CONFLICT (id) DO UPDATE
	SET
		name = EXCLUDED.name,
		lat = EXCLUDED.lat,
		lon = EXCLUDED.lon,
		ebikes_available = EXCLUDED.ebikes_available,
		bike_docks_available = EXCLUDED.bike_docks_available,
		ebikes = EXCLUDED.ebikes,
		created_at = now() at time zone 'utc';

-- name: GetStations :many
select
	id,
	name,
	lat,
	lon,
	ebikes_available,
	bike_docks_available,
	ebikes,
	ST_MakePoint(lon, lat) <-> ST_MakePoint( sqlc.arg(lon)::float, sqlc.arg(lat)::float ) AS distance,
	created_at
from stations
where ebikes_available > 0
order by distance asc
limit 10;
