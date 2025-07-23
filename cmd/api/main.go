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

	"github.com/lieberdev/go-rest-template/internal/database"
	"github.com/lieberdev/go-rest-template/internal/mailer"
)

type config struct {
	port int
	env string
	cors struct {
		allowedOrigins []string
	}
	smtp mailer.Config
	db database.Config
}

type application struct {
	config     config
	logger     *slog.Logger
	mailer     *mailer.Mailer
	models     database.Models
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
	flag.StringVar(&cfg.smtp.Host, "smtp-host", "", "SMTP host (required)")
	flag.IntVar(&cfg.smtp.Port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.Username, "smtp-username", "", "SMTP username (required)")
	flag.StringVar(&cfg.smtp.Password, "smtp-password", "", "SMTP password (required)")
	flag.StringVar(&cfg.smtp.Sender, "smtp-sender", "", "SMTP sender (required)")
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

	// Check if required flags are set
	missing := []string{}
	if cfg.smtp.Host == "" { missing = append(missing, "--smtp-host") }
	if cfg.smtp.Username == "" { missing = append(missing, "--smtp-username") }
	if cfg.smtp.Password == "" { missing = append(missing, "--smtp-password") }
	if cfg.smtp.Sender == "" { missing = append(missing, "--smtp-sender") }
	if len(missing) > 0 {
		slog.Error("missing required flags: " + strings.Join(missing, ", "))
		os.Exit(1)
	}

	db, err := database.Init(&cfg.db)
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
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stat()
	}))
	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: database.NewModels(db),
		mailer: mailer,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
