#!/usr/bin/env bash

set -euo pipefail

./devlib/buildgo.py --all
./devlib/buildvscode.py
./devlib/package.py --all
