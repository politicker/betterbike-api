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
	cp -f com.pumpfactory.betterbike.plist ~/Library/LaunchAgents/

load:
	launchctl load ~/Library/LaunchAgents/com.pumpfactory.betterbike.plist

unload:
	launchctl unload ~/Library/LaunchAgents/com.pumpfactory.betterbike.plist


deploy:
	rm /usr/local/bin/betterbike && \
	go build -o /usr/local/bin/betterbike && \
	just unload && \
	just copy-plist && \
	just load

dump:
	./bin/dump-database

