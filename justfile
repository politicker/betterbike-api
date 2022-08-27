set dotenv-load

gen:
	sqlc generate

resetdb:
	dropdb citibike-dev && createdb citibike-dev && psql citibike-dev < db/schema.sql
