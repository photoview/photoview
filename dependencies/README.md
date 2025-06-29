# Dependencies

This directory contains scripts and Dockerfile to build third-party dependencies for Photoview. It is not intended for end-users or runtime execution.

However, technically, it is runnable for debugging and investigative purposes.
It contains a single archive file, `artifacts.tar.gz`, that bundles all dependencies.

You can unpack it with:

```bash
tar -xzf artifacts.tar.gz [-C /desired/location]
```
