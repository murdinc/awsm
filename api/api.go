package api

import (
	"net/http"
	"os"

	"github.com/goware/cors"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/skratchdot/open-golang/open"
)

// StartDashboard Starts the Dashboard port 8081
func StartDashboard() {
	open.Start("http://localhost:8081")
	StartAPI()
}

// StartAPI Starts the API listener on port 8081
func StartAPI() {
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
	})

	r.Use(cors.Handler)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	//r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Route("/dashboard", func(r chi.Router) {
			r.Route("/widgets", func(r chi.Router) {
				r.Get("/", getWidgets)
				r.Get("/events", getEvents)
				r.Get("/feed/:feedName", getFeed)
				r.Get("/options", getWidgetOptions)
				r.Get("/names", getWidgetNames)
				r.Get("/name/:widgetName", getWidgetByName)
				r.Put("/name/:widgetName", putWidget)
				r.Delete("/name/:widgetName", deleteWidget)
			})
		})
		r.Route("/assets", func(r chi.Router) {
			r.Route("/:assetType", func(r chi.Router) {
				r.Get("/", getAssets)
			})
		})
		r.Route("/classes", func(r chi.Router) {
			r.Get("/export", exportClasses)
			//r.Get("/import", importClasses) // TODO
			r.Route("/:classType", func(r chi.Router) {
				r.Get("/", getClasses)
				r.Get("/options", getClassOptions)
				r.Get("/names", getClassNames)
				r.Get("/name/:className", getClassByName)
				r.Put("/name/:className", putClass)
				r.Delete("/name/:className", deleteClass)
			})
		})
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.RequestURI()
		src := "/usr/local/awsmDashboard"

		if uri == "/" {
			src += "/index.html"
		} else if _, err := os.Stat(src + uri); os.IsNotExist(err) {
			src += "/index.html"
		} else {
			src += uri
		}

		// return file
		http.ServeFile(w, r, src)

	})

	http.ListenAndServe(":8081", r)
}
