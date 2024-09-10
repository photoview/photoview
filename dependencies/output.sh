#!/bin/sh
set -eu

cd /output
tar czfv /artifacts.tar.gz *
