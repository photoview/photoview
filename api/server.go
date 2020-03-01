package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/gorilla/mux"

	"github.com/joho/godotenv"

	"github.com/viktorstrate/photoview/api/database"
	"github.com/viktorstrate/photoview/api/graphql/auth"
	"github.com/viktorstrate/photoview/api/routes"
	"github.com/viktorstrate/photoview/api/server"

	"github.com/99designs/gqlgen/handler"
	photoview_graphql "github.com/viktorstrate/photoview/api/graphql"
	"github.com/viktorstrate/photoview/api/graphql/resolvers"
)

const defaultPort = "4001"

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

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

	endpointURL, err := url.Parse(os.Getenv("API_ENDPOINT"))
	if err != nil {
		log.Println("WARN: Environment variable API_ENDPOINT not specified")
		endpointURL, _ = url.Parse("/")
	}

	endpointRouter := rootRouter.PathPrefix(endpointURL.Path).Subrouter()

	if devMode {
		endpointRouter.Handle("/api", handler.Playground("GraphQL playground", path.Join(endpointURL.Path, "/graphql")))
	} else {
		endpointRouter.HandleFunc("/api", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("photoview api endpoint"))
		})
	}

	endpointRouter.Handle("/api/graphql",
		handler.GraphQL(photoview_graphql.NewExecutableSchema(graphqlConfig),
			handler.IntrospectionEnabled(devMode),
			handler.WebsocketUpgrader(server.WebsocketUpgrader(devMode)),
			handler.WebsocketInitFunc(auth.AuthWebsocketInit(db)),
		),
	)

	photoRouter := endpointRouter.PathPrefix("/api/photo").Subrouter()
	routes.RegisterPhotoRoutes(db, photoRouter)

	spa := spaHandler{staticPath: "/ui", indexPath: "index.html"}
	endpointRouter.PathPrefix("/").Handler(spa)

	if devMode {
		log.Printf("🚀 Graphql playground ready at %s", endpointURL.String())
	} else {
		log.Printf("Photoview API endpoint available at %s", endpointURL.String())
	}

	log.Fatal(http.ListenAndServe(":"+port, rootRouter))
}
