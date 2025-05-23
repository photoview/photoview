#!/bin/bash
set -euo pipefail

cd /output
tar czfv /artifacts.tar.gz *
