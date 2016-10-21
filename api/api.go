package api

import (
	"net/http"

	"github.com/goware/cors"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

// StartAPI Starts the API listener on port 8081
func StartAPI() {
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
	})

	r.Use(cors.Handler)
	//r.Use(middleware.RequestID)
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Route("/assets", func(r chi.Router) {
			r.Route("/:assetType", func(r chi.Router) {
				r.Get("/", getAssets)
			})
		})
		r.Route("/classes", func(r chi.Router) {
			r.Route("/:classType", func(r chi.Router) {
				r.Get("/", getClasses)
				r.Get("/names", getClassNames)
				r.Get("/name/:className", getClassByName)
				r.Put("/name/:className", putClass)
				r.Delete("/name/:className", deleteClass)
			})
		})
	})

	http.ListenAndServe(":8081", r)
}
