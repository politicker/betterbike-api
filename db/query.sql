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
		ebikes = EXCLUDED.ebikes;

-- name: GetStations :many
select
	id,
	name,
	lat,
	lon,
	ebikes_available,
	bike_docks_available,
	ebikes,
	ST_MakePoint(lat, lon) <-> ST_MakePoint( sqlc.arg(lat), sqlc.arg(lon) ) AS distance
from stations
order by distance asc
limit 10;
