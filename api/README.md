# GRANDstack Starter - GraphQL API


## Quick Start

Install dependencies:

```
npm install
```

Start the GraphQL service:

```
npm start
```

This will start the GraphQL service (by default on localhost:4000) where you can issue GraphQL requests or access GraphQL Playground in the browser:

![GraphQL Playground](img/graphql-playground.png)

## Configure

Set your Neo4j connection string and credentials in `.env`. For example:

*.env*

```
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=letmein
```

## Deployment

You can deploy to any service that hosts Node.js apps, but [Zeit Now](https://zeit.co/now) is a great easy to use service for hosting your app that has an easy to use free plan for small projects. 

To deploy your GraphQL service on Zeit Now, first install [Now Desktop](https://zeit.co/download) - you'll need to provide an email address. Then run

```
now
```

to deploy your GraphQL service on Zeit Now. Once deployed you'll be given a fresh URL that represents the current state of your application where you can access your GraphQL endpoint and GgraphQL Playgound. For example: https://grand-stack-starter-api-pqdeodpvok.now.sh/

## Seeding The Database

Optionally you can seed the GraphQL service by executing mutations that will write sample data to the database:

```
npm run seedDb
```