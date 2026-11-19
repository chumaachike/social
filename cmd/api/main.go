package main

import (
	"log"
	"time"

	"github.com/chumaachike/social/internal/db"
	"github.com/chumaachike/social/internal/env"
	"github.com/chumaachike/social/internal/repository"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		env:  "development",

		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetDuration("DB_MAX_IDLE_TIME", time.Duration(15*time.Minute)),
		},
	}

	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("db connection pool established")

	repo := repository.NewRepository(db)

	app := &application{
		config:     cfg,
		repository: repo,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
