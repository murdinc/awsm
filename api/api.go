package api

import (
	"net/http"
	"os"

	"github.com/goware/cors"
	"github.com/murdinc/terminal"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/skratchdot/open-golang/open"
)

// StartAPI Starts the API listener on port 8081
func StartAPI(withDashboard bool) error {
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

	if withDashboard {
		src := "/usr/local/awsmDashboard"

		_, err := os.Stat(src)
		if os.IsNotExist(err) {
			// TODO install the dashboard automatically
			terminal.Notice("No awsm dashboard found!")
			terminal.Notice("Please install the dashboard by running the following command:")
			terminal.Notice("curl -s http://dl.sudoba.sh/get/awsmDashboard | sh")
			return err
		}

		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			uri := r.URL.RequestURI()

			//fmt.Println(r.URL.Path)

			if uri == "/" {
				uri = "/index.html"
			} else if _, err := os.Stat(src + uri); os.IsNotExist(err) {
				uri = "/index.html"
			}

			terminal.Information("Serving file: " + src + uri)

			// return file
			http.ServeFile(w, r, src+uri)

		})

		open.Start("http://localhost:8081")
	}

	return http.ListenAndServe(":8081", r)
}
