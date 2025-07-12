package main

import (
	"expvar"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	router.Use(app.metrics)
	router.Use(middleware.Recoverer)
	router.Use(app.Logger)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.config.cors.allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	router.Use(httprate.Limit(
		20,
		1*time.Minute,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			app.tooManyRequests(w, r)
		}),
	))
	router.Use(app.authenticate)

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	router.Get("/healthcheck", app.healthcheckHandler)
	router.Get("/debug/vars", expvar.Handler().ServeHTTP)

	router.Group(func(router chi.Router) {
		router.Use(httprate.Limit(
			3,
			1*time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				app.tooManyRequests(w, r)
			}),
		))

		router.Post("/users", app.registerUserHandler)
		router.Put("/users/activated", app.activateUserHandler)
		router.Put("/users/password", app.updateUserPasswordHandler)

		router.Post("/tokens/authentication", app.createAuthenticationTokenHandler)
		router.Post("/tokens/activation", app.createActivationTokenHandler)
		router.Post("/tokens/password-reset", app.createPasswordResetTokenHandler)
	}

	return router
}
