# PhotoView

![screenshot](/screenshots/main-window.png)

## Demo site
Visit: [photos.qpqp.dk](http://photos.qpqp.dk/)

Username: __demo__
Password: __demo__

## Aim of the project

The aim of this project is to make a simple and user-friendly photo gallery application,
that is easy to host on a personal server, to easily view the photos located on that server.

## Main features
> The software is still in early development, and many of the following features, have not been implemented yet.

- **Closely tied to the file system**. The website presents the images found on the local filesystem of the server, directories are mapped to albums.
- **User management**. Each user is created along with a path on the local filesystem, photos within that path can be accessed by that user.
- **Photo sharing**. Photos and albums can easily be shared with other users or publicly with a unique URL.
- **Made for photography**. The website is ment as a way to present photographies, and thus supports **RAW** file formats, and **EXIF** parsing.

## Getting started - Setup with Docker

> This section describes how to get PhotoView up and running on your server with Docker.
> Make sure you have Docker and docker-compose installed and running on your server

1. Clone this repository by executing

```bash
$ git clone https://github.com/viktorstrate/photoview
$ cd photoview
```

2. Duplicate `docker-compose.proxy-example.yml` and name the new file `docker-compose.yml`
3. Edit `docker-compose.yml`, find the comments starting with `Change This:`, and change the values, to properly match your setup.
4. Start the server by running the following command, inside the `photoview` directory

```bash
$ docker-compose up -d
```

If the endpoint or the port hasn't been changed in the `docker-compose.yml` file, PhotoView can now be accessed at http://localhost:8080

### Initial Setup

If everything is setup correctly, you should be presented with an initial setup wizard, when accessing the website the first time.

![Initial setup](/screenshots/initial-setup.png)

Enter a new username and password.

For the photo path, enter the path in the docker container where your photos are located.
This can be set from the `docker-compose.yml` file under `api` -> `volumes`.
The default location is `/photos`

A new admin user will be created, with access to the photos located at the path provided under the initial setup.

The photos will have to be scanned for the photos to show up, you can force a scan, by navigating to `Settings` and clicking on `Scan All`

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

#### Install dependencies

```bash
(cd ./ui && npm install)
(cd ./api && npm install)
```

#### Start API server

```bash
cd ./api && npm start
```

### [`/ui`](./ui)

This will start the GraphQL API in the foreground, so in another terminal session start the UI development server:

#### Start UI server

```bash
cd ./ui && npm start
```

The site can now be accessed at [localhost:1234](http://localhost:1234).
And the graphql playground at [localhost:4001/graphql](http://localhost:4001/graphql)
