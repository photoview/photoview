import { typeDefs, resolvers } from "./graphql-schema";
import { ApolloServer, gql } from "apollo-server";
import { v1 as neo4j } from "neo4j-driver";
import dotenv from "dotenv";

dotenv.config();

const driver = neo4j.driver(
  process.env.NEO4J_URI || "bolt://localhost:7687",
  neo4j.auth.basic(
    process.env.NEO4J_USER || "neo4j",
    process.env.NEO4J_PASSWORD || "neo4j"
  )
);

const server = new ApolloServer({
  typeDefs,
  resolvers,
  context: { driver }
});

server.listen().then(({ url }) => {
  console.log(`GraphQL API read at ${url}`);
});