import { neo4jgraphql } from "neo4j-graphql-js";

export const typeDefs = `
type User {
    name: ID!
}

type Query {
    users: [User]
}
`;

export const resolvers = {
  Query: {
    users: neo4jgraphql
  }
};
