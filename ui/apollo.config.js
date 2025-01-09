// apollo.config.js
const { readdirSync, readFileSync, writeFileSync, mkdirSync } = require('node:fs');
const path = require('node:path');

const schemasFolder = path.join(__dirname, '..', 'api', 'graphql', 'resolvers');

var completeSchema;
try {
   completeSchema = readdirSync(schemasFolder)
    .filter(x => x.endsWith('.graphql'))
    .map(x => {
      const filePath = path.join(schemasFolder, x);
      return readFileSync(filePath, 'utf-8');
    })
    .join('\n\n');
} catch (error) {
  console.error('Failed to generate schema:', error);
  process.exit(1);
}

const outputPath = path.join(__dirname, '.cache', 'schema.graphql');

try {
  mkdirSync(path.dirname(outputPath), { recursive: true });
  writeFileSync(outputPath, completeSchema, { mode: 0o644 });
} catch (error) {
  console.error('Failed to write schema file:', error);
  process.exit(1);
}

module.exports = {
  client: {
    service: {
      name: 'photoview',
      localSchemaFile: outputPath,
    },
  },
}
