<img src="./screenshots/photoview-logo.svg" height="92px" alt="photoview logo" />

[![License](https://img.shields.io/github/license/viktorstrate/photoview)](./LICENSE.md)
[![GitHub contributors](https://img.shields.io/github/contributors/viktorstrate/photoview)](https://github.com/viktorstrate/photoview/graphs/contributors)
[![Docker Pulls](https://img.shields.io/docker/pulls/viktorstrate/photoview)](https://hub.docker.com/r/viktorstrate/photoview)
[![Docker Build Status](https://img.shields.io/github/workflow/status/viktorstrate/photoview/Docker%20builds?label=docker%20build)](https://hub.docker.com/r/viktorstrate/photoview/)

![screenshot](./screenshots/timeline.png)

Photoview is a simple and user-friendly photo gallery that can easily be installed on personal servers.
It's made for photographers and aims to provide an easy and fast way to navigate directories, with thousands of high resolution photos.

> If you have questions regarding setup or development,
feel free to join the Discord server https://discord.gg/jQ392948u9

## Demo site

Visit https://photos.qpqp.dk/

Username: **demo**
Password: **demo**

## Contents

- [Demo site](#demo-site)
- [Main features](#main-features)
- [Why yet another self-hosted photo gallery](#why-yet-another-self-hosted-photo-gallery)
- [Getting started - Setup with Docker](#getting-started---setup-with-docker)
- [Environment Variables](#available-environment-variables)
- [Setup development environment](#setup-development-environment)

## Main features

- **Closely tied to the file system**. The website presents the images found on the local filesystem of the server, directories are mapped to albums.
- **User management**. Each user is created along with a path on the local filesystem, photos within that path can be accessed by that user.
- **Sharing**. Albums, as well as individual media, can easily be shared with a public link, the link can optinally be password protected.
- **Made for photography**. Photoview is built with photographers in mind, and thus supports **RAW** file formats, and **EXIF** parsing.
- **Video support**. Many common video formats are supported. Videos will automatically be optimized for web.
- **Face recognition**. Faces will automatically be detected in photos, and photos of the same person will be grouped together.
- **Performant**. Thumbnails are automatically generated and photos first load when they are visible on the screen. In full screen, thumbnails are displayed until the high resolution image has been fully loaded.
- **Secure**. All media resources are protected with a cookie-token, all passwords are properly hashed, and the API uses a strict [CORS policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).

## Why yet another self-hosted photo gallery

There exists a lot of open-source self-hosted photo galleries already. Here are some, just to mention a few.

- [Piwigo](https://github.com/Piwigo/Piwigo)
- [Ownphoto](https://github.com/hooram/ownphotos)
- [Photoprism](https://github.com/photoprism/photoprism)
- [Lychee](https://github.com/LycheeOrg/Lychee)

So why another one?
I love taking photos, and I store all of them on my local fileserver.
This is great because I can organize my photos directly on the filesystem so it's easy to move them or take backups. I want to be able to control where and how the photos are stored.

The problem is however that RAW images are extremely tedious to navigate from a fileserver, even over the local network.

My server holds a lot of old family pictures, that I would like my family to have access to as well.
And some of the pictures I would like to easily be able to share with other people without the hassle of them having to make an account first.

Thus I need a solution that can do the following:

- A scan based approach that automatically organises my photos
- Support RAW and EXIF parsing
- Have support for multiple users and ways to share albums and photos also publicly
- Be simple and fast to use

All of the photo galleries can do a lot of what I need, but no single one can do it all.

## Getting started - Setup with Docker

> This section describes how to get Photoview up and running on your server with Docker.
> Make sure you have Docker and docker-compose installed and running on your server

1. Make a new `docker-compose.yml` file on your computer, and copy the content of [docker-compose.example.yml](/docker-compose.example.yml) to the new file.
2. Edit `docker-compose.yml`, find the comments starting with `Change This:`, and change the values, to properly match your setup. If you are just testing locally, you don't have to change anything.
3. Start the server by running the following command

```bash
$ docker-compose up -d
```

If the endpoint or the port hasn't been changed in the `docker-compose.yml` file, Photoview can now be accessed at http://localhost:8000

### Initial Setup

If everything is setup correctly, you should be presented with an initial setup wizard, when accessing the website the first time.

![Initial setup](./screenshots/initial-setup.png)

Enter a new username and password.

For the photo path, enter the path in the docker container where your photos are located.
This can be set from the `docker-compose.yml` file under `api` -> `volumes`.
The default location is `/photos`

A new admin user will be created, with access to the photos located at the path provided under the initial setup.

The photos will have to be scanned before they show up, you can start a scan manually, by navigating to `Settings` and clicking on `Scan All`

## Available Environment Variables
> This section describes all available environment variables that can be configured.

Required |Variable | Default | Notes 
---------|---------|---------|------
|:eight_pointed_black_star: |*_Database related_* | :eight_pointed_black_star: |:eight_pointed_black_star:     
:heavy_check_mark: | PHOTOVIEW_MYSQL_URL | "" | The URL of the MYSQL database to connect to. Formatting [here](https://github.com/go-sql-driver/mysql#dsn-data-source-name)
:heavy_check_mark: | PHOTOVIEW_DATABASE_DRIVER | "" | Driver to use for database- ie. "mysql"
:heavy_check_mark: | PHOTOVIEW_LISTEN_PORT | "" | Port that photoview listens on
:heavy_check_mark: | PHOTOVIEW_LISTEN_IP | "" | IP that photoview listens on 

#### cache related

Required |Variable | Default | Notes 
---------|---------|---------|------
:heavy_check_mark: | PHOTOVIEW_PUBLIC_ENDPOINT | "" | URL where endpoint can be accessed. Specify FQDN with proper protocol- ie. http://localhost:8000 or https://photos.example.com
:white_check_mark: | PHOTOVIEW_MEDIA_CACHE | ./photo_cache | Filepath for where to store generated media such as thumbnails and optimized videos
:white_check_mark: | PHOTOVIEW_MAPBOX_TOKEN | "" | To enable map related features, you need to create a mapbox token. A token can be generated for free at https://account.mapbox.com/access-tokens/ It's a good idea to limit the scope of the token to your own domain, to prevent others from using it.

## Setup development environment

### Local setup

1. Install a local mysql server, and make a new database
2. Rename `/api/example.env` to `.env` and update the `MYSQL_URL` field
3. Rename `/ui/example.env` to `.env`

### Start API server

Make sure [golang](https://golang.org/) is installed.

Some C libraries are needed to compile the API, see [go-face requirements](https://github.com/Kagami/go-face#requirements) for more details.
They can be installed as shown below:

```sh
# Ubuntu
sudo apt-get install libdlib-dev libblas-dev liblapack-dev libjpeg-turbo8-dev
# Debian
sudo apt-get install libdlib-dev libblas-dev liblapack-dev libjpeg62-turbo-dev
# macOS
brew install dlib

```

Then run the following commands:

```bash
cd ./api && go run server.go
```

### Start UI server

Make sure [node](https://nodejs.org/en/) is installed.
In a new terminal window run the following commands:

```bash
cd ./ui && npm start
```

The site can now be accessed at [localhost:1234](http://localhost:1234).
And the graphql playground at [localhost:4001](http://localhost:4001)
