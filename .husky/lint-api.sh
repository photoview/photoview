#!/bin/sh

gofiles=$(git diff --cached --name-only --diff-filter=ACM | grep '.go$')
[ -z "$gofiles" ] && exit 0

# Automatically format go code, exit on error
echo "Formatting staged go files"
gofmt -w $gofiles
