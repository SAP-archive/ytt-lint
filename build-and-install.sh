#!/usr/bin/env bash

set -euo pipefail

CODE=code
if command -v codium >/dev/null ; then
    CODE=codium
fi

EXTENSION_VERSION="$(jq -r ".version" vscode/package.json)"
OS="$(uname | tr "[:upper:]" "[:lower:]")"
export OS
export EXTENSION_VERSION

./devlib/buildgo.py

./devlib/buildvscode.py

./devlib/package.py

set -x
# install extension
$CODE --install-extension "out/ytt-lint-$EXTENSION_VERSION-$OS.vsix"
