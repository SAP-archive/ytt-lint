#!/usr/bin/env bash

set -euo pipefail

CODE=code
if command -v codium >/dev/null ; then
    CODE=codium
fi

EXTENSION_VERSION="$(jq -r ".version" vscode/package.json)"

set -x

./build-all.sh
cd vscode/
cp ../out/* bin/
vsce package
$CODE --install-extension "ytt-lint-$EXTENSION_VERSION.vsix"