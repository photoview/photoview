# PhotoView

> NOTE: This software should not be used in production yet, since it is in early development, and still contains security holes.

![screenshot](/screenshot.png)

## Aim of the project

The aim of this project is to make a simple and user-friendly photo gallery application,
that is easy to host on a personal server, to easily view the photos located on that server.

## Main features
> The software is still in early development, and many of the following features, have not been implemented yet.

- **Closely tied to the file system**. The website presents the images found on the local filesystem of the server, directories are mapped to albums.
- **User management**. Each user is created along with a path on the local filesystem, photos within that path can be accessed by that user.
- **Photo sharing**. Photos and albums can easily be shared with other users or publicly with a unique URL.
- **Made for photography**. The website is ment as a way to present photographies, and thus supports **RAW** file formats, and **EXIF** parsing.

## Setup development environment
> This projected is based of the [GrandStack](https://grandstack.io/) starter project.

### Local setup
1. [Download Neo4j Desktop](https://neo4j.com/download/)
2. Install and open Neo4j Desktop.
3. Create a new DB by clicking "New Graph", and clicking "create local graph".
4. Set password to "letmein" (as suggested by `api/.env`), and click "Create".
5. Make sure that the default credentials in `api/.env` are used. Leave them as follows: `NEO4J_URI=bolt://localhost:7687 NEO4J_USER=neo4j NEO4J_PASSWORD=letmein`
6.  Click "Manage".
7. Click "Plugins".
8. Find "APOC" and click "Install".
9. Click the "play" button at the top of left the screen, which should start the server. _(screenshot 2)_
10. Wait until it says "RUNNING".
11. Proceed forward with the rest of the tutorial.

### [`/api`](./api)

*Install dependencies*

```
(cd ./ui && npm install)
(cd ./api && npm install)
```

*Start API server*
```
cd ./api && npm start
```

![](api/img/graphql-playground.png)

### [`/ui`](./ui)

This will start the GraphQL API in the foreground, so in another terminal session start the UI development server:

*Start UI server*
```
cd ./ui && npm start
```

The site can now be accessed at [localhost:1234](http://localhost:1234).
And the graphql playground at [localhost:4001/graphql](http://localhost:4001/graphql)

## Docker Compose

> Not written yet