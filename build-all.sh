#!/usr/bin/env bash

mkdir -p out

GOOS=linux   go build -o out/ytt-lint-linux  github.com/k14s/ytt/cmd/ytt-lint/
GOOS=darwin  go build -o out/ytt-lint-darwin github.com/k14s/ytt/cmd/ytt-lint/
GOOS=windows go build -o out/ytt-lint-win32  github.com/k14s/ytt/cmd/ytt-lint/
