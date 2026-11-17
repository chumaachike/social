package main

import (
	"log"

	"github.com/chumaachike/social/internal/env"
	"github.com/chumaachike/social/internal/repository"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	repo := repository.NewRepository(nil)

	app := &application{
		config:     cfg,
		repository: repo,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
