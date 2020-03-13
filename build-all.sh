#!/usr/bin/env bash

mkdir -p out

GOOS=linux   go build -ldflags="-s -w" -o out/ytt-lint-linux  ./cmd/ytt-lint/
GOOS=darwin  go build -ldflags="-s -w" -o out/ytt-lint-darwin ./cmd/ytt-lint/
GOOS=windows go build -ldflags="-s -w" -o out/ytt-lint-win32  ./cmd/ytt-lint/
