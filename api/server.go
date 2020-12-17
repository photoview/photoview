package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/routes"
	"github.com/photoview/photoview/api/scanner"
	"github.com/photoview/photoview/api/server"
	"github.com/photoview/photoview/api/utils"

	"github.com/99designs/gqlgen/handler"
	photoview_graphql "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/resolvers"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	devMode := os.Getenv("DEVELOPMENT") == "1"

	db, err := database.SetupDatabase()
	if err != nil {
		log.Panicf("Could not connect to database: %s\n", err)
	}
	defer db.Close()

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

	scanner.InitializeExecutableWorkers()

	rootRouter := mux.NewRouter()

	rootRouter.Use(auth.Middleware(db))
	rootRouter.Use(server.LoggingMiddleware)
	rootRouter.Use(server.CORSMiddleware(devMode))

	graphqlResolver := resolvers.Resolver{Database: db}
	graphqlDirective := photoview_graphql.DirectiveRoot{}
	graphqlDirective.IsAdmin = photoview_graphql.IsAdmin(db)

	graphqlConfig := photoview_graphql.Config{
		Resolvers:  &graphqlResolver,
		Directives: graphqlDirective,
	}

	apiListenUrl := utils.ApiListenUrl()

	endpointRouter := rootRouter.PathPrefix(apiListenUrl.Path).Subrouter()

	if devMode {
		endpointRouter.Handle("/", handler.Playground("GraphQL playground", path.Join(apiListenUrl.Path, "/graphql")))
	} else {
		endpointRouter.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("photoview api endpoint"))
		})
	}

	endpointRouter.Handle("/graphql",
		handler.GraphQL(photoview_graphql.NewExecutableSchema(graphqlConfig),
			handler.IntrospectionEnabled(devMode),
			handler.WebsocketUpgrader(server.WebsocketUpgrader(devMode)),
			handler.WebsocketInitFunc(auth.AuthWebsocketInit(db)),
		),
	)

	photoRouter := endpointRouter.PathPrefix("/photo").Subrouter()
	routes.RegisterPhotoRoutes(db, photoRouter)

	videoRouter := endpointRouter.PathPrefix("/video").Subrouter()
	routes.RegisterVideoRoutes(db, videoRouter)

	shouldServeUI := os.Getenv("SERVE_UI") == "1"

	if shouldServeUI {
		spa := routes.NewSpaHandler("/ui", "index.html")
		rootRouter.PathPrefix("/").Handler(spa)
	}

	if devMode {
		log.Printf("🚀 Graphql playground ready at %s\n", apiListenUrl.String())
	} else {
		log.Printf("Photoview API endpoint listening at %s\n", apiListenUrl.String())

		uiEndpoint := utils.UiEndpointUrl()
		apiEndpoint := utils.ApiEndpointUrl()

		log.Printf("Photoview API public endpoint ready at %s\n", apiEndpoint.String())
		log.Printf("Photoview UI public endpoint ready at %s\n", uiEndpoint.String())

		if !shouldServeUI {
			log.Printf("Notice: UI is not served by the the api (SERVE_UI=0)")
		}

	}

	log.Panic(http.ListenAndServe(":"+apiListenUrl.Port(), handlers.CompressHandler(rootRouter)))
}
