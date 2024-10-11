package main

import (
	"log"
	"net/http"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/auth"
	graphql_endpoint "github.com/photoview/photoview/api/graphql/endpoint"
	"github.com/photoview/photoview/api/routes"
	"github.com/photoview/photoview/api/scanner/exif"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/periodic_scanner"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/photoview/photoview/api/server"
	"github.com/photoview/photoview/api/utils"

	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {

	log.Println("Starting Photoview...")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	devMode := utils.DevelopmentMode()

	db, err := database.SetupDatabase()
	if err != nil {
		log.Panicf("Could not connect to database: %s\n", err)
	}

	// Migrate database
	if err := database.MigrateDatabase(db); err != nil {
		log.Panicf("Could not migrate database: %s\n", err)
	}

	if err := scanner_queue.InitializeScannerQueue(db); err != nil {
		log.Panicf("Could not initialize scanner queue: %s\n", err)
	}

	if err := periodic_scanner.InitializePeriodicScanner(db); err != nil {
		log.Panicf("Could not initialize periodic scanner: %s", err)
	}

	executable_worker.InitializeExecutableWorkers()

	exif.InitializeEXIFParser()

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		log.Panicf("Could not initialize face detector: %s\n", err)
	}

	rootRouter := mux.NewRouter()

	rootRouter.Use(dataloader.Middleware(db))
	rootRouter.Use(auth.Middleware(db))
	rootRouter.Use(server.LoggingMiddleware)
	rootRouter.Use(server.CORSMiddleware(devMode))

	apiListenURL := utils.ApiListenUrl()

	endpointRouter := rootRouter.PathPrefix(apiListenURL.Path).Subrouter()

	if devMode {
		endpointRouter.Handle("/", playground.Handler("GraphQL playground", path.Join(apiListenURL.Path, "/graphql")))
	} else {
		endpointRouter.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("photoview api endpoint"))
		})
	}

	endpointRouter.Handle("/graphql", graphql_endpoint.GraphqlEndpoint(db))

	photoRouter := endpointRouter.PathPrefix("/photo").Subrouter()
	routes.RegisterPhotoRoutes(db, photoRouter)

	videoRouter := endpointRouter.PathPrefix("/video").Subrouter()
	routes.RegisterVideoRoutes(db, videoRouter)

	downloadsRouter := endpointRouter.PathPrefix("/download").Subrouter()
	routes.RegisterDownloadRoutes(db, downloadsRouter)

	shouldServeUI := utils.ShouldServeUI()

	if shouldServeUI {
		spa := routes.NewSpaHandler(utils.UIPath(), "index.html")
		rootRouter.PathPrefix("/").Handler(spa)
	}

	if devMode {
		log.Printf("🚀 Graphql playground ready at %s\n", apiListenURL.String())
	} else {
		log.Printf("Photoview API endpoint listening at %s\n", apiListenURL.String())

		apiEndpoint := utils.ApiEndpointUrl()
		log.Printf("Photoview API public endpoint ready at %s\n", apiEndpoint.String())

		logUIendpointURL()

		if !shouldServeUI {
			log.Printf("Notice: UI is not served by the the api (%s=0)", utils.EnvServeUI.GetName())
		}

	}

	log.Panic(http.ListenAndServe(apiListenURL.Host, handlers.CompressHandler(rootRouter)))
}

func logUIendpointURL() {
	if uiEndpoint := utils.UiEndpointUrl(); uiEndpoint != nil {
		log.Printf("Photoview UI public endpoint ready at %s\n", uiEndpoint.String())
	} else {
		log.Println("Photoview UI public endpoint ready at /")
	}
}
