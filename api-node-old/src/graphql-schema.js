import fs from 'fs-extra'
import path from 'path'
import { makeAugmentedSchema } from 'neo4j-graphql-js'
import _ from 'lodash'

import usersResolver from './resolvers/users'
import scannerResolver from './resolvers/scanner'
import photosResolver from './resolvers/photos'
import siteInfoResolver from './resolvers/siteInfo'
import sharingResolver from './resolvers/sharing'

const resolvers = [
  usersResolver,
  scannerResolver,
  photosResolver,
  siteInfoResolver,
  sharingResolver,
]

const typeDefs = fs
  .readFileSync(
    process.env.GRAPHQL_SCHEMA || path.join(__dirname, 'schema.graphql')
  )
  .toString('utf-8')

let productionExcludes = []

// if (process.env.PRODUCTION == true) {
//   productionExcludes = [
//     'ScannerResult',
//     'AuthorizeResult',
//     'PhotoURL',
//     'SiteInfo',
//     'User',
//     'Album',
//     'PhotoEXIF',
//     'Photo',
//     'ShareToken',
//     'Result',
//   ]
// }

const schema = makeAugmentedSchema({
  typeDefs,
  config: {
    auth: {
      isAuthenticated: true,
      hasRole: true,
    },
    mutation: false,
    query: {
      exclude: [...productionExcludes],
    },
  },
  resolvers: resolvers.reduce((prev, curr) => _.merge(prev, curr), {}),
})

export default schema
