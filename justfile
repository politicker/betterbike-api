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

load-mac:
	launchctl load ~/Library/LaunchAgents/com.betterbike.plist

unload-mac:
	launchctl unload ~/Library/LaunchAgents/com.betterbike.plist

load:
    systemctl --user stop betterbike.service || true
    systemctl --user disable betterbike.service || true
    sudo rm -f $HOME/.config/systemd/user/betterbike.service
    sudo cp daemon/betterbike.service $HOME/.config/systemd/user/betterbike.service
    systemctl --user daemon-reload
    systemctl --user enable betterbike.service
    systemctl --user start betterbike.service

sysstatus:
	systemctl --user status betterbike.service

syslogs:
	journalctl --user-unit=betterbike.service --no-pager

syslogstail:
	journalctl --user-unit=betterbike.service

deploy-mac:
	rm betterbike && \
	go build -o betterbike && \
	just unload && \
	just copy-plist && \
	just load

dump:
	./bin/dump-database

psql:
	psql -h localhost -p 5433 -U $USER -d betterbike-api

