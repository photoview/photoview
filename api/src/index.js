import { typeDefs, resolvers } from "./graphql-schema";
import { ApolloServer, gql } from "apollo-server";

const server = new ApolloServer({
  typeDefs,
  resolvers
});

server.listen().then(({ url }) => {
  console.log(`GraphQL API read at ${url}`);
});
