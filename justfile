set dotenv-load

setup:
	brew bundle && \
	createdb betterbike

gen:
	sqlc generate

resetdb:
	dropdb betterbike && createdb betterbike && psql betterbike < internal/db/schema.sql

build:
	go build -o betterbike

run:
	go run .

copy-plist:
	cp -f daemon/com.betterbike.plist ~/Library/LaunchAgents/

load:
	launchctl load ~/Library/LaunchAgents/com.betterbike.plist

unload:
	launchctl unload ~/Library/LaunchAgents/com.betterbike.plist

deploy:
	rm betterbike-api && \
	go build -o betterbike-api && \
	just unload && \
	just copy-plist && \
	just load

dump:
	./bin/dump-database

