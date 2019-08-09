module.exports = {
  client: {
    service: {
      name: 'graphql endpoint',
      url: process.env.GRAPHQL_ENDPOINT || 'http://localhost:4001/graphql',
      skipSSLValidation: true,
    },
  },
}
