package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/apoliticker/citibike/db"
	_ "github.com/lib/pq"
)

var connectionString string
var queries *db.Queries

func init() {
	connectionString = os.Getenv("DATABASE_CONN_STRING")

	database, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic("failed to connect to database")
	}

	queries = db.New(database)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// TODO: Pass cancellable context to poller and server
	// TODO: Pass logger to poller and server

	poller := NewPoller(queries, 1*time.Minute)
	go poller.Start()

	srv := NewServer(port, queries)
	srv.Start()
}
