<img src="./screenshots/photoview-logo.svg" height="92px" alt="photoview logo" />

[![License](https://img.shields.io/github/license/kkovaletp/photoview)](./LICENSE.txt)
[![GitHub contributors](https://img.shields.io/github/contributors/kkovaletp/photoview)](https://github.com/kkovaletp/photoview/graphs/contributors)
[![Docker Pulls](https://img.shields.io/docker/pulls/kkoval/photoview)](https://hub.docker.com/r/kkoval/photoview)
[![Docker builds](https://github.com/kkovaletp/photoview/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/kkovaletp/photoview/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/kkovaletp/photoview/branch/master/graph/badge.svg?token=ENP02P5NYS)](https://codecov.io/gh/kkovaletp/photoview)
<a href="https://scan.coverity.com/projects/kkovaletp-photoview">
  <img alt="Coverity Scan Build Status"
       src="https://scan.coverity.com/projects/30259/badge.svg"/>
</a>

[![SonarCloud](https://sonarcloud.io/images/project_badges/sonarcloud-white.svg)](https://sonarcloud.io/summary/new_code?id=kkovaletp_photoview)

**Backend**

[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-api&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-api)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-api&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-api)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-api&metric=bugs)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-api)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-api&metric=coverage)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-api)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-api&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-api)

**Frontend**

[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-ui&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-ui)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-ui&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-ui)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-ui&metric=bugs)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-ui)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-ui&metric=coverage)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-ui)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-ui&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-ui)

**Docker**

[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-docker&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-docker)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-docker&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-docker)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-docker&metric=bugs)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-docker)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=kostiantyn-github_photoview-docker&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=kostiantyn-github_photoview-docker)


![screenshot](./screenshots/timeline.png)

**This is a forked repository from the [photoview/photoview](https://github.com/photoview/photoview).** 
My vision of the product development strategy is different from the Photoview owner and the maintainer of the upstream repo, so I decided to fork and go my way.

**Photoview** is a simple and user-friendly photo gallery that's made for photographers and aims to provide an easy and fast way to navigate directories, with thousands of high-resolution photos.

You configure Photoview to look for photos and videos within a directory on your file system. The scanner automatically picks up your media and starts to generate thumbnail images to make browsing super fast.

When your media has been scanned, they show up on the website, organized in the same way as on the filesystem.

> If you have questions regarding setup or development,
feel free to start or join a discussion in this repo

## Terms of use

By using this project or its source code, for any purpose and in any shape or form, you grant your **implicit agreement** to all of the following statements:

- You unequivocally condemn Russia and its military aggression against Ukraine;
- You recognize that Russia is an occupant that unlawfully invaded a sovereign state;
- You agree that [Russia is a terrorist state](https://www.europarl.europa.eu/doceo/document/RC-9-2022-0482_EN.html);
- You fully support Ukraine's territorial integrity, including its claims over [temporarily occupied territories](https://en.wikipedia.org/wiki/Russian-occupied_territories_of_Ukraine);
- You reject false narratives perpetuated by Russian state propaganda.

To learn more about the war and how you can help, [click here](https://war.ukraine.ua/).

Glory to Ukraine! 🇺🇦

## Contents

- [Terms of use](#Terms-of-use)
- [Main features](#main-features)
- [Supported Platforms](#supported-platforms)
- [Why yet another self-hosted photo gallery](#why-yet-another-self-hosted-photo-gallery)
- [Getting started — Setup with Docker](#getting-started--setup-with-docker)
- [Advanced setup](#advanced-setup)
- [Contributing](#contributing)
- [Set up development environment](#set-up-development-environment)

## Main features

- **Closely tied to the file system**. The website presents the images found on the local filesystem of the server; directories are mapped to albums.
- **User management**. Each user is created along with a path on the local filesystem, photos within that path can be accessed by that user.
- **Sharing**. Albums, as well as individual media, can easily be shared with a public link, the link can optionally be password protected.
- **Made for photography**. Photoview is built with photographers in mind, and thus supports **RAW** file formats, and **EXIF** parsing.
- **Video support**. Many common video formats are supported. Videos will automatically be optimized for web.
- **Face recognition**. Faces will automatically be detected in photos, and photos of the same person will be grouped together.
- **Performant**. Thumbnails are automatically generated and photos first load when they are visible on the screen. In full screen, thumbnails are displayed until the high-resolution image has been fully loaded.
- **Secure**. All media resources are protected with a cookie-token, all passwords are properly hashed, and the API uses a strict [CORS policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).

## Supported platforms

- [Docker](https://hub.docker.com/r/kkoval/photoview/) - recommended and preferred
- Debian, Ubuntu and similar Linux distros
- Fedora Linux
- [Arch Linux Aur](https://aur.archlinux.org/packages/photoview)
- [Unraid](https://forums.unraid.net/topic/103028-support-photoview-corneliousjd-repo/)
- EmbassyOS: [announcement](https://start9labs.medium.com/new-service-photoview-72ee681b2ff0), [repo](https://github.com/Start9Labs/embassyos-photoview-wrapper)
- [YunoHost](https://github.com/YunoHost-Apps/photoview_ynh)

## Why yet another self-hosted photo gallery

There exists a lot of open-source self-hosted photo galleries already. Here are some, just to mention a few.

- [Piwigo](https://github.com/Piwigo/Piwigo)
- [LibrePhotos](https://github.com/LibrePhotos/librephotos)
- [Photoprism](https://github.com/photoprism/photoprism)
- [Lychee](https://github.com/LycheeOrg/Lychee)

So why another one?
I love taking photos, and I store all of them on my local fileserver.
This is great because I can organize my photos directly on the filesystem, so it's easy to move them or take backups. I want to be able to control where and how the photos are stored.

The problem is, however, that RAW images are extremely tedious to navigate from a fileserver, even over the local network.

My server holds a lot of old family pictures that I would like my family to have access to as well.
And some of the pictures I would like to easily be able to share with other people without the hassle of them having to make an account first.

Thus, I need a solution that can do the following:

- A scan-based approach that automatically organises my photos
- Support RAW and EXIF parsing
- Have support for multiple users and ways to share albums and photos also publicly
- Be straightforward and fast to use

All the photo galleries can do a lot of what I need, but no single one can do it all.

## Getting started — Setup with Docker

> This section describes how to get Photoview up and running on your server with Docker.
> Make sure you have Docker and docker-compose installed and running on your server.
> `make` should be installed as well if you'd like to use provided `Makefile`, which is optional (see step 4 for more details).
> `7zz` should be installed in case, you'd like to use it in scope of the backup scenario instead of the default .tar.xz format. Read the comment in the `Makefile`, located on top of the `backup` section for more details.

1. Download the content of the `docker-compose example` folder to the folder on your server, where you expect to host the Photoview internal data (database and cache files).
   
   Please note that this folder contains 2 versions of the docker-compose file:
   - `docker-compose.example.yml` - the fully-functional and recommended for the most cases config
   - `docker-compose.minimal.example.yml` - the minimal and simple config for those, who find the previous one too complex and difficult to understand and manage
   
   When downloading files, you need to choose only one of them.
2. Rename downloaded files and remove the `example` from their names (so, you need to have `.env`, `docker-compose.yml`, and `Makefile` files). If you choose the `docker-compose.minimal.example.yml` on previous step, make sure to rename it to the `docker-compose.yml`.
3. Open these files in a text editor and read them. Modify where needed according to the documentation comments to properly match your setup. There are comments of 2 types: those, starting with `##`, are explanations and examples, which should not be uncommented; those, starting with `#`, are optional or alternative configuration parts, which might be uncommented in certain circumstances, described in corresponding explanations. It is better to go through the files in the next order: `.env`, `docker-compose.yml`, and `Makefile`.
4. Make sure that your media library's root folder and all the files and subfolders are readable and searchable by other users: run the next command (or corresponding sequence of commands from the `Makefile`):

   ```bash
   make readable
   ```

   If command(s) return `Permission denied` error, run them under the user, owning corresponding files and folders. Alternatively, run them adding `sudo ` before the command: this will switch the execution context to `root` user and ask for the root password. You have to have permission to run `sudo` in the system.

   If you don't want to give required permissions to `others` group for your files, alternatively, you can:

   - create a group on your host with GID=999 and make all the files and folders inside volumes of the `photoview` service being owned by this group; then set the appropriate permissions to the `group` section.
   - create on your host a group with GID=999 and a user in this group with UID=999; then change the ownership of all the files and folders inside volumes of the `photoview` service to this user; then set the appropriate permissions to the `user` section.

   If you configured other mounts with media files from other locations on the host (like HOST_PHOTOVIEW_MEDIA_FAMILY or anything else), you need to run the same commands, as in the `Makefile` `readable` target, for each media root folder on your host manually: copy each command to your shell and replace the variable with the absolute path to an additional media root folder without the trailing `/`. Run both commands for each additional root folder.
5. In case, you don't have `make` installed in your system or don't want to use it for the Photoview management activities, you could use the same commands from the `Makefile` and run them in your shell directly, or create your own scripts. Make sure to apply or replace the variables from your `.env` first in this case. `Makefile` is provided just for your convenience and simplicity, but is optional.
6. Start the server by running the following command (or corresponding sequence of commands from the `Makefile`):

   ```bash
   make all
   ```

If the endpoint or the port hasn't been changed in the `docker-compose.yml` file, Photoview can now be accessed at http://localhost:8000

### Initial Setup

If everything is set up correctly, you should be presented with an initial setup wizard when accessing the website the first time.

![Initial setup](./screenshots/initial-setup.png)

Enter a new username and password.

For the photo path, enter the path inside the docker container where your photos are located.
This can be set from the `docker-compose.yml` file under `photoview` -> `volumes`.
The default location is `/photos`.

A new admin user will be created, with access to the photos located at the path provided under the initial setup.

The photos will have to be scanned before they show up, you can start a scan manually, by navigating to `Settings` and clicking on `Scan All`

## Advanced setup

I suggest securing the Photoview instance before exposing it outside your local network: even while it provides read-only access to your media gallery and has basic user authentication functionality, it is not enough to protect your private media from malicious actors on the Internet.

Possible ways of securing a self-hosted service might be (but not limited to):

1. Configure a **Firewall** on your local network's gateway and allow only the intended type of incoming traffic to pass.
2. Use **VPN** to provide external access to local services.
3. Setting up a **Reverse proxy** in front of the service and forwarding all the traffic through it, exposing HTTPS port with strong certificate and cipher suites to the Internet. This could be one of the next products or something else that you prefer:
   - [Traefic Proxy](https://doc.traefik.io/traefik/)
   - [NGinx Proxy Manager](https://nginxproxymanager.com/guide/)
   - [Cloudflare Gateway](https://www.cloudflare.com/zero-trust/products/gateway/)
4. Configure an external **Multi-Factor Authentication** service to manage authentication for your service (part of Cloudflare services, but you can choose anything else).
5. Configure **Web Application Firewall** to protect from common web exploits like SQL injection, cross-site scripting, and cross-site forgery requests (part of Cloudflare services, but you can choose anything else).
6. Use **Content Delivery Network** as an additional level of DDoS prevention: it can securely cache your media and let it be accessible from a wide list of servers on the Internet (part of Cloudflare services, but you can choose anything else).
7. Configure a **Rate Limit** of allowed number of requests from a user during specified time range to protect against DDoS attacks.
8. Set up an **Intrusion Detection/Prevention System** to monitor network traffic for suspicious activity and issue alerts when such activity is discovered.

Setting up and configuring of all these protections depends on and requires a lot of info about your local network and self-hosted services. Based on this info, the configuration flow and resulting services architecture might differ a lot between cases. That is why in the scope of this project, we can only provide you with this high-level list of possible ways of webservice protection. You'll need to investigate them, find the best combination and configuration for your case, and take responsibility to configure everything in the correct and consistent way. We cannot provide you support for such highly secured setups, as a lot of things might work differently because of security limitations.

## Contributing

🎉 First off, thanks for your interest in contribution! 🎉

This project is a result of hard work, and it's great to see you interested in contributing. 
Contributions are not just about code — you can help in many ways!

Before you start, please take a moment to read our [Contributing guide](./CONTRIBUTING.md). 
It includes information on our code of conduct, the process for submitting pull requests, and more.

Remember, every contribution counts. Let's make this project better together! 💪

## Set up development environment

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
sudo add-apt-repository ppa:strukturag/libheif
sudo add-apt-repository ppa:strukturag/libde265
sudo apt-get update
sudo apt-get install libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg-turbo8-dev libheif-dev
# Debian
sudo apt-get install libdlib-dev libblas-dev libatlas-base-dev liblapack-dev libjpeg62-turbo-dev libheif-dev
# macOS
brew install dlib libheif

```

Then run the following commands:

```bash
cd ./api
go install
go run server.go
```

### Start UI server

Make sure [node](https://nodejs.org/en/) is installed.
In a new terminal window run the following commands:

```bash
cd ./ui
npm install
npm start
```

The site can now be accessed at [localhost:1234](http://localhost:1234).
And the graphql playground at [localhost:4001](http://localhost:4001)
