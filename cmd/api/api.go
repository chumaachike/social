package main

import (
	"net/http"
	"time"

	"github.com/chumaachike/social/docs"
	"github.com/chumaachike/social/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config     config
	repository repository.Repository
	logger     *zap.SugaredLogger
}

type config struct {
	addr   string
	db     dbConfig
	env    string
	apiURL string
	mail   mailConfig
}

type mailConfig struct {
	exp time.Duration
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has time out and further processing should be stopped.
	r.Use(middleware.Timeout(time.Minute))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/v1/swagger/doc.json"),
		))

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPost)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware) //allow routes under /postID to use postID context
				r.Get("/", app.getPost)
				r.Delete("/", app.deletePost)
				r.Patch("/", app.updatePost)
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUser)
				r.Put("/follow", app.followUser)
				r.Put("/unfollow", app.unfollowUser)
			})
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeed)
			})
		})
		// Public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("Server starting", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}
