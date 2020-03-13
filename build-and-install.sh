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

set -x

# Build go
./build-all.sh

# Build extension
pushd vscode/
    vsce package
popd

# Pack executable to extension
# zip "vscode/ytt-lint-$EXTENSION_VERSION.vsix" "out/ytt-lint-$OS" --output-file "out/ytt-lint-$EXTENSION_VERSION-$OS.vsix"

cp "vscode/ytt-lint-$EXTENSION_VERSION.vsix" "out/ytt-lint-$EXTENSION_VERSION-$OS.vsix"

python3 <<PY
import zipfile
import os

OS = os.environ["OS"]
EXTENSION_VERSION = os.environ["EXTENSION_VERSION"]

zip = zipfile.ZipFile("out/ytt-lint-%s-%s.vsix" % (EXTENSION_VERSION, OS), 'a', compression=zipfile.ZIP_DEFLATED)
zip.write("out/ytt-lint-%s" % OS, "extension/bin/ytt-lint")
zip.close()
PY

# install extension
$CODE --install-extension "out/ytt-lint-$EXTENSION_VERSION-$OS.vsix"