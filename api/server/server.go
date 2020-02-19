package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/viktorstrate/photoview/api/database"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/routes"

	"github.com/99designs/gqlgen/handler"
	photoview_graphql "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/resolvers"
)

const defaultPort = "4001"

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	devMode := os.Getenv("DEVELOPMENT") == "1"

	port := os.Getenv("API_LISTEN_PORT")
	if port == "" {
		port = defaultPort
	}

	db := database.SetupDatabase()
	defer db.Close()

	// Migrate database
	if err := database.MigrateDatabase(db); err != nil {
		log.Fatalf("Could not migrate database: %s\n", err)
	}

	router := chi.NewRouter()
	router.Use(auth.Middleware(db))

	router.Use(middleware.Logger)

	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4001", "http://localhost:1234", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	graphqlResolver := resolvers.Resolver{Database: db}
	graphqlDirective := photoview_graphql.DirectiveRoot{}
	graphqlDirective.IsAdmin = photoview_graphql.IsAdmin(db)

	graphqlConfig := photoview_graphql.Config{
		Resolvers:  &graphqlResolver,
		Directives: graphqlDirective,
	}

	endpointURL, err := url.Parse(os.Getenv("API_ENDPOINT"))
	if err != nil {
		log.Println("WARN: Environment variable API_ENDPOINT not specified")
		endpointURL, _ = url.Parse("/")
	}

	router.Route(endpointURL.Path, func(router chi.Router) {
		if devMode {
			router.Handle("/", handler.Playground("GraphQL playground", path.Join(endpointURL.Path, "/graphql")))
		} else {
			router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("photoview api endpoint"))
			})
		}

		router.Handle("/graphql", handler.GraphQL(photoview_graphql.NewExecutableSchema(graphqlConfig), handler.IntrospectionEnabled(devMode)))

		router.Mount("/photo", routes.PhotoRoutes(db))
	})

	if devMode {
		log.Printf("🚀 Graphql playground ready at %s", endpointURL.String())
	} else {
		log.Printf("Photoview API endpoint available at %s", endpointURL.String())
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
}
