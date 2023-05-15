set dotenv-load

setup:
	brew bundle && \
	createdb citibike-dev

gen:
	sqlc generate

resetdb:
	dropdb citibike-dev && createdb citibike-dev && psql citibike-dev < db/schema.sql

build:
	go build -o betterbike

run:
	go run .

copy-plist:
	cp -f com.pumpfactory.betterbike.plist ~/Library/LaunchAgents/

load:
	launchctl load ~/Library/LaunchAgents/com.pumpfactory.betterbike.plist

unload:
	launchctl unload ~/Library/LaunchAgents/com.pumpfactory.betterbike.plist
