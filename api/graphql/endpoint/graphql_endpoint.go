package graphql_endpoint

import (
	"time"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	photoview_graphql "github.com/kkovaletp/photoview/api/graphql"
	"github.com/kkovaletp/photoview/api/graphql/resolvers"
	"github.com/kkovaletp/photoview/api/server"
	"github.com/kkovaletp/photoview/api/utils"
	"gorm.io/gorm"
)

func GraphqlEndpoint(db *gorm.DB) *graphql_handler.Server {
	graphqlResolver := resolvers.NewRootResolver(db)
	graphqlDirective := photoview_graphql.DirectiveRoot{}
	graphqlDirective.IsAdmin = photoview_graphql.IsAdmin
	graphqlDirective.IsAuthorized = photoview_graphql.IsAuthorized

	graphqlConfig := photoview_graphql.Config{
		Resolvers:  &graphqlResolver,
		Directives: graphqlDirective,
	}

	graphqlServer := graphql_handler.New(photoview_graphql.NewExecutableSchema(graphqlConfig))
	graphqlServer.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader:              server.WebsocketUpgrader(utils.DevelopmentMode()),
	})
	graphqlServer.AddTransport(transport.Options{})
	graphqlServer.AddTransport(transport.GET{})
	graphqlServer.AddTransport(transport.POST{})
	graphqlServer.AddTransport(transport.MultipartForm{})

	graphqlServer.SetQueryCache(lru.New(1000))

	graphqlServer.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	if utils.DevelopmentMode() {
		graphqlServer.Use(extension.Introspection{})
	}

	return graphqlServer
}
