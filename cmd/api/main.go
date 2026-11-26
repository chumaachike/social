package main

import (
	"time"

	"github.com/chumaachike/social/internal/db"
	"github.com/chumaachike/social/internal/env"
	"github.com/chumaachike/social/internal/repository"
	"go.uber.org/zap"
)

const version = "0.0.1"

// @title						Social API
// @version					0.0.1
// @description				API for a social networking application.
// @BasePath					/v1
// @in							header
// @name						Authorization
// @description				Enter "Bearer" followed by your JWT token.
func main() {
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		env:    "development",
		mail: mailConfig{
			exp: time.Hour * 24 * 3,
		},

		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetDuration("DB_MAX_IDLE_TIME", time.Duration(15*time.Minute)),
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("db connection pool established")

	repo := repository.NewRepository(db)

	app := &application{
		config:     cfg,
		repository: repo,
		logger:     logger,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
