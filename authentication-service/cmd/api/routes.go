package main

import (
	chi2 "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"net/http"
)

func (app *Config) routes() http.Handler {
	chi := chi2.NewRouter()

	chi.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTION"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content_Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	chi.Use(middleware.Heartbeat("/ping"))

	chi.Post("/authenticate", app.Authenticate)

	return chi
}
