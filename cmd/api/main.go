package main

import (
	"expvar"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/lieberdev/go-rest-template/internal/data"
	"github.com/lieberdev/go-rest-template/internal/db"
	"github.com/lieberdev/go-rest-template/internal/mailer"
)

type config struct {
	port int
	env string
	cors struct {
		allowedOrigins []string
	}
	smtp mailer.Config
	db db.Config
}

type application struct {
	config     config
	logger     *slog.Logger
	models     data.Models
	mailer     mailer.Mailer
	waitgroup  sync.WaitGroup
}

func main() {
	var cfg config

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Server
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Database
	flag.StringVar(&cfg.db.Dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(
		&cfg.db.MaxIdleTime,
		"db-max-idle-time",
		"15m",
		"PostgreSQL max connection idle time",
	)
	// Mailer
	flag.StringVar(&cfg.smtp.Host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.Port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.Username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.smtp.Password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.Sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "SMTP sender")
	// CORS
	flag.Func(
		"cors-allowed-origins",
		"Allowed CORS origins (space separated)",
		func(val string) error {
			cfg.cors.allowedOrigins = strings.Fields(val)
			return nil
		},
	)
	flag.Parse()

	db, err := db.Init(&cfg.db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	mailer, err := mailer.Init(cfg.smtp, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("mailer connection established")

	// Publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
