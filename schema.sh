#!/usr/bin/env bash

set -euo pipefail

mkdir -p schema
cd schema

pip install openapi2jsonschema

schema='https://raw.githubusercontent.com/kubernetes/kubernetes/v1.17.0/api/openapi-spec/swagger.json'

openapi2jsonschema -o "k8s" --stand-alone "${schema}"

cd k8s

for schema in *.json ; do
    target=$(jq -r 'select(."x-kubernetes-group-version-kind" != null) | ."x-kubernetes-group-version-kind"[0] | "\(if .group == "" then "core" else .group end)/\(.version)/\(.kind | ascii_downcase).json"' "$schema")

    if [ -n "$target" ] ; then
        mkdir -p "$(dirname "$target")"
        mv "$schema" "$target"
    else
        rm "$schema"
    fi
done