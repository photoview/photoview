package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/auth"
	graphql_endpoint "github.com/photoview/photoview/api/graphql/endpoint"
	"github.com/photoview/photoview/api/routes"
	"github.com/photoview/photoview/api/scanner/externaltools/exif"
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
		log.Println("No .env file found. If Photoview runs in Docker, this is expected and correct.")
	}

	terminateWorkers := executable_worker.Initialize()
	defer terminateWorkers()

	devMode := utils.DevelopmentMode()

	db, err := database.SetupDatabase()
	if err != nil {
		log.Panicf("Could not connect to database: %s\n", err)
	}

	// Migrate database
	if err := database.MigrateDatabase(db); err != nil {
		log.Panicf("Could not migrate database: %s\n", err)
	}

	exifCleanup, err := exif.Initialize()
	if err != nil {
		log.Panicf("Could not initialize exif parser: %s", err)
	}
	defer exifCleanup()

	if err := scanner_queue.InitializeScannerQueue(db); err != nil {
		log.Panicf("Could not initialize scanner queue: %s\n", err)
	}

	if err := periodic_scanner.InitializePeriodicScanner(db); err != nil {
		log.Panicf("Could not initialize periodic scanner: %s", err)
	}

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
		endpointRouter.Handle("/", playground.Handler("GraphQL playground", path.Join(apiListenURL.Path, "graphql")))
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
			log.Printf("Notice: UI is not served by the API (%s=0)", utils.EnvServeUI.GetName())
		}

	}

	srv := &http.Server{
		Addr:    apiListenURL.Host,
		Handler: handlers.CompressHandler(rootRouter),
	}

	setupGracefulShutdown(srv)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Panicf("HTTP server failed: %s", err)
	}
}

func setupGracefulShutdown(svr *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down Photoview...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute) // Wait for 1m to shutdown
		defer cancel()

		// Shutdown scanners in correct order
		periodic_scanner.ShutdownPeriodicScanner()
		scanner_queue.CloseScannerQueue()

		if err := svr.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %s", err)
		} else {
			log.Println("Shutdown complete")
		}
	}()
}

func logUIendpointURL() {
	if uiEndpoint := utils.UiEndpointUrl(); uiEndpoint != nil {
		log.Printf("Photoview UI public endpoint ready at %s\n", uiEndpoint.String())
	} else {
		log.Println("Photoview UI public endpoint ready at /")
	}
}
