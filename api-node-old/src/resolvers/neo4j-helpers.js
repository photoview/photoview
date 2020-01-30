import { cypherQuery } from 'neo4j-graphql-js'

// Helper functions, that makes it easier to manipulate neo4j-graphql-js translations

export function replaceMatch({ root, args, ctx, info }, match) {
  let query = cypherQuery(args, ctx, info)[0]

  query = query.substr(query.indexOf(')') + 1)
  query = match + query

  return query
}
