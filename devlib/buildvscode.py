#!/usr/bin/env python3

import subprocess
import os
import util

def build() -> None:
    build_env = os.environ.copy()
    subprocess.check_call(['vsce', 'package'], env=build_env, shell=False, cwd=util.getextensiondir())

if __name__ == "__main__":
    build()
