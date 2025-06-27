# Dependencies

This directory contains scripts and Dockerfile to build third-party dependencies for Photoview. It is not intended for end-users or runtime execution.

To locally debug this image you need to run 2 simple steps:

1. Build the image locally and tag it as `dependencies:self`. For example:

    `docker buildx build --pull --tag dependencies:self .`

2. Build the debug image, which you then can run and investigate the content:

    `docker buildx build --pull --tag deps-debug --file ./Dockerfile-debug .`

Now you can run the debug image by:

`docker run --rm -it deps-debug /bin/bash`

Alternatively, you can build it up to the `final-assembly` target and run it:

```bash
docker buildx build --security-opt seccomp=unconfined --pull --target final-assembly --tag dependencies:self .
docker run --rm -it dependencies:self /bin/bash
```
