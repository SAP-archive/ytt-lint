#!/usr/bin/env bash

set -euo pipefail

mkdir -p schema
cd schema
mkdir -p k8s

pip install openapi2jsonschema

for version in 1.14.0 1.15.0 1.16.0 1.17.0 1.18.0 ; do
    schema="https://raw.githubusercontent.com/kubernetes/kubernetes/v$version/api/openapi-spec/swagger.json"
    openapi2jsonschema -o "k8s-$version" --stand-alone "${schema}"

    pushd "k8s-$version"
    for schema in *.json ; do
        target=$(jq -r 'select(."x-kubernetes-group-version-kind" != null) | ."x-kubernetes-group-version-kind"[0] | "../k8s/\(if .group == "" then "core" else .group end)/\(.version)/\(.kind | ascii_downcase).json"' "$schema")

        if [ -n "$target" ] ; then
            mkdir -p "$(dirname "$target")"
            mv "$schema" "$target"
        else
            rm "$schema"
        fi
    done
    popd
    rm -rf "k8s-$version"
done
