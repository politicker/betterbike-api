set dotenv-load

setup:
	brew bundle && \
	createdb citibike-dev

gen:
	sqlc generate

resetdb:
	dropdb citibike-dev && createdb citibike-dev && psql citibike-dev < db/schema.sql
