CREATE EXTENSION if not exists postgis;

create table
    if not exists stations (
        id text primary key,
        name text not null,
        lat float not null,
        lon float not null,
        ebikes_available int not null default 0,
        bike_docks_available int not null default 0,
        ebikes jsonb not null default '{}',
        created_at timestamp not null default now ()
    );

create table
    if not exists stations_timeseries (
        id text not null,
        name text not null,
        lat float not null,
        lon float not null,
        bikes_available int not null default 0,
        ebikes_available int not null default 0,
        bike_docks_available int not null default 0,
        last_updated_ms bigint not null,
        is_offline boolean not null default false,
        created_at timestamp not null default now (),
        PRIMARY KEY (id, last_updated_ms)
    );
