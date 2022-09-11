package main

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"os"
	"time"

	"github.com/apoliticker/citibike/db"
	_ "github.com/lib/pq"
	_ "github.com/politicker/zap-sink-datadog"
)

var databaseURL string
var queries *db.Queries
var logger *zap.Logger

func init() {
	var config zap.Config
	var err error

	if os.Getenv("LOG_TO_DATADOG") == "true" {
		config = zap.NewProductionConfig()
		config.OutputPaths = []string{"dd://us5.datadoghq.com/betterbikes-api?tags=env:production"}
	} else {
		config = zap.NewDevelopmentConfig()
	}

	logger, err = config.Build()
	if err != nil {
		panic(err)
	}

	databaseURL = os.Getenv("DATABASE_URL")
	logger.Info("connecting to db", zap.String("databaseURL", databaseURL))

	databaseURL = fmt.Sprintf("%s?sslmode=disable", databaseURL)

	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database")
	}

	queries = db.New(database)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// TODO: Pass cancellable context to poller and server
	poller := NewPoller(queries, logger, 1*time.Minute)
	go poller.Start()

	srv := NewServer(port, queries, logger)
	srv.Start()

	logger.Sync()
}
