gen:
	pggen gen go \
		-postgres-connection "dbname=citibike-dev" \
		--schema-glob db/schema.sql \
		--query-glob db/query.sql


resetdb:
	dropdb citibike-dev && createdb citibike-dev && psql citibike-dev < db/schema.sql
