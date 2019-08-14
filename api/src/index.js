import { ApolloServer } from 'apollo-server-express'
import express from 'express'
import bodyParser from 'body-parser'
import cors from 'cors'
import { v1 as neo4j } from 'neo4j-driver'
import dotenv from 'dotenv'
import http from 'http'
import PhotoScanner from './scanner/Scanner'
import _ from 'lodash'
import config from './config'

import { getUserFromToken, getTokenFromBearer } from './token'

// set environment variables from ../.env
dotenv.config()

const app = express()
app.use(bodyParser.json())
app.use(cors())

/*
 * Create a Neo4j driver instance to connect to the database
 * using credentials specified as environment variables
 * with fallback to defaults
 */
const driver = neo4j.driver(
  process.env.NEO4J_URI || 'bolt://localhost:7687',
  neo4j.auth.basic(
    process.env.NEO4J_USER || 'neo4j',
    process.env.NEO4J_PASSWORD || 'letmein'
  )
)

const scanner = new PhotoScanner(driver)

// Every 4th hour
setInterval(scanner.scanAll, 1000 * 60 * 60 * 4)

// Specify port and path for GraphQL endpoint
const graphPath = '/graphql'

const endpointUrl = new URL(config.host)
// endpointUrl.port = process.env.GRAPHQL_LISTEN_PORT || 4001

/*
 * Create a new ApolloServer instance, serving the GraphQL schema
 * created using makeAugmentedSchema above and injecting the Neo4j driver
 * instance into the context object so it is available in the
 * generated resolvers to connect to the database.
 */

import schema from './graphql-schema'

const server = new ApolloServer({
  context: async function({ req }) {
    let user = null
    let token = null

    if (req && req.headers.authorization) {
      token = getTokenFromBearer(req.headers.authorization)
      user = await getUserFromToken(token, driver)
    }

    return {
      ...req,
      driver,
      scanner,
      user,
      token,
      endpoint: endpointUrl.toString(),
    }
  },
  schema,
  introspection: true,
  playground: !process.env.PRODUCTION,
  subscriptions: {
    onConnect: async (connectionParams, webSocket) => {
      const token = getTokenFromBearer(connectionParams.Authorization)
      const user = await getUserFromToken(token, driver)

      return {
        token,
        user,
      }
    },
  },
})

server.applyMiddleware({ app, path: graphPath })

import loadImageRoutes from './routes/images'

loadImageRoutes({ app, driver, scanner })

const httpServer = http.createServer(app)
server.installSubscriptionHandlers(httpServer)

httpServer.listen(
  { port: process.env.GRAPHQL_LISTEN_PORT, path: graphPath },
  () => {
    console.log(
      `🚀 GraphQL endpoint ready at ${new URL(server.graphqlPath, endpointUrl)}`
    )

    let subscriptionUrl = new URL(endpointUrl)
    subscriptionUrl.protocol = 'ws'

    console.log(
      `🚀 Subscriptions ready at ${new URL(
        server.subscriptionsPath,
        endpointUrl
      )}`
    )
  }
)
