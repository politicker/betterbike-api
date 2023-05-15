-- InsertStation inserts citibike station data into the database.
-- name: InsertStation :exec
insert into stations (id,
                      name,
                      lat,
                      lon,
                      ebikes_available,
                      bike_docks_available,
                      ebikes,
                      created_at)
values ($1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        now() at time zone 'utc') ON CONFLICT (id) DO
UPDATE
    SET
        name = EXCLUDED.name,
    lat = EXCLUDED.lat,
    lon = EXCLUDED.lon,
    ebikes_available = EXCLUDED.ebikes_available,
    bike_docks_available = EXCLUDED.bike_docks_available,
    ebikes = EXCLUDED.ebikes,
    created_at = now() at time zone 'utc';

-- InsertStationTimeseries appends station data to the timeseries table.
-- name: InsertStationTimeseries :exec
insert into stations_timeseries (id,
                                 name,
                                 lat,
                                 lon,
																 bikes_available,
                                 ebikes_available,
                                 bike_docks_available,
                                 last_updated_ms,
																 is_offline)
values ($1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
				$9) ON CONFLICT (id, last_updated_ms) DO NOTHING;

-- name: GetStations :many
select id,
       name,
       lat,
       lon,
       ebikes_available,
       bike_docks_available,
       ebikes,
       (
           ST_DistanceSphere(
                   ST_MakePoint(lon, lat),
                   ST_MakePoint(sqlc.arg(lon)::float, sqlc.arg(lat)::float)
               )
           )::float AS distance, created_at
from stations
where ebikes_available > 0
  and (
          ST_DistanceSphere(
                  ST_MakePoint(lon, lat),
                  ST_MakePoint(sqlc.arg(lon)::float, sqlc.arg(lat)::float)
              )
          ) < 3000 -- approx. 2 miles
order by distance limit 10;
