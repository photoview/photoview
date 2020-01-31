package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/viktorstrate/photoview/api/database"

	"github.com/99designs/gqlgen/handler"
	photoview_graphql "github.com/viktorstrate/photoview/api/graphql"
)

const defaultPort = "4001"

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = defaultPort
	}

	db := database.SetupDatabase()
	defer db.Close()

	// Migrate database
	if err := database.MigrateDatabase(db); err != nil {
		log.Fatalf("Could not migrate database: %s\n", err)
	}

	graphqlResolver := photoview_graphql.Resolver{Database: db}

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/query", handler.GraphQL(photoview_graphql.NewExecutableSchema(photoview_graphql.Config{Resolvers: &graphqlResolver})))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
