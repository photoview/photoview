// apollo.config.js
const { readdirSync, readFileSync, writeFileSync } = require('fs');

const schemasFolder = __dirname + '/../api/graphql/resolvers';
const completeSchema = readdirSync(schemasFolder)
  .filter(x => x.endsWith('.graphql'))
  .map(x => readFileSync(`${schemasFolder}/${x}`, 'utf-8'))
  .join('\n\n');

writeFileSync('/tmp/schema.graphql', completeSchema);

module.exports = {
  client: {
    service: {
      name: 'photoview',
      localSchemaFile: '/tmp/schema.graphql',
    },
  },
}
