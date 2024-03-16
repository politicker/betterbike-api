package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"github.com/politicker/betterbike-api/internal/db"
	"github.com/politicker/betterbike-api/internal/web"
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
	if databaseURL == "" {
		logger.Fatal("missing DATABASE_URL")
	}

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

	if os.Getenv("SENTRY_DSN") != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_DSN"),
			TracesSampleRate: 1.0,
			Environment:      os.Getenv("SENTRY_ENV"),
			Release:          os.Getenv("RELEASE_NAME"),
		})
		if err != nil {
			logger.Fatal("sentry.Init failed", zap.Error(err))
		}
	}

	// TODO: Pass cancellable context to poller and server
	logger.Info("starting poller")
	poller := NewPoller(queries, logger.With(zap.String("context", "poller")), 1*time.Minute)
	go poller.Start()

	logger.Info("starting server")
	srv := NewServer(port, queries, logger.With(zap.String("context", "server")))
	go srv.Start()

	logger.Info("starting html server", zap.String("port", port))
	wsrv := web.NewWeb(context.Background(), logger.With(zap.String("context", "html-server")), queries, "8001")
	wsrv.Start()

	logger.Sync()
}
