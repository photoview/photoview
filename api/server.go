package main

import (
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/routes"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/scanner/exif"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/server"
	"github.com/photoview/photoview/api/utils"

	"github.com/99designs/gqlgen/handler"
	photoview_graphql "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/resolvers"
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

	if err := scanner.InitializeScannerQueue(db); err != nil {
		log.Panicf("Could not initialize scanner queue: %s\n", err)
	}

	if err := scanner.InitializePeriodicScanner(db); err != nil {
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

	graphqlResolver := resolvers.Resolver{Database: db}
	graphqlDirective := photoview_graphql.DirectiveRoot{}
	graphqlDirective.IsAdmin = photoview_graphql.IsAdmin
	graphqlDirective.IsAuthorized = photoview_graphql.IsAuthorized

	graphqlConfig := photoview_graphql.Config{
		Resolvers:  &graphqlResolver,
		Directives: graphqlDirective,
	}

	apiListenURL := utils.ApiListenUrl()

	endpointRouter := rootRouter.PathPrefix(apiListenURL.Path).Subrouter()

	if devMode {
		endpointRouter.Handle("/", handler.Playground("GraphQL playground", path.Join(apiListenURL.Path, "/graphql")))
	} else {
		endpointRouter.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("photoview api endpoint"))
		})
	}

	endpointRouter.Handle("/graphql",
		handler.GraphQL(photoview_graphql.NewExecutableSchema(graphqlConfig),
			handler.IntrospectionEnabled(devMode),
			handler.WebsocketUpgrader(server.WebsocketUpgrader(devMode)),
			handler.WebsocketKeepAliveDuration(time.Second*10),
			handler.WebsocketInitFunc(auth.AuthWebsocketInit(db)),
		),
	)

	photoRouter := endpointRouter.PathPrefix("/photo").Subrouter()
	routes.RegisterPhotoRoutes(db, photoRouter)

	videoRouter := endpointRouter.PathPrefix("/video").Subrouter()
	routes.RegisterVideoRoutes(db, videoRouter)

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

		if uiEndpoint := utils.UiEndpointUrl(); uiEndpoint != nil {
			log.Printf("Photoview UI public endpoint ready at %s\n", uiEndpoint.String())
		} else {
			log.Println("Photoview UI public endpoint ready at /")
		}

		if !shouldServeUI {
			log.Printf("Notice: UI is not served by the the api (%s=0)", utils.EnvServeUI.GetName())
		}

	}

	log.Panic(http.ListenAndServe(":"+apiListenURL.Port(), handlers.CompressHandler(rootRouter)))
}
