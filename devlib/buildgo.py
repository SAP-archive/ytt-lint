#!/usr/bin/env python3

import subprocess
import platform
import os
import sys
import util

def build(target_os:str = "native") -> None:
    if target_os == "native":
        target_os = platform.system().lower()
    print("Building for %s" % target_os)
    build_env = os.environ.copy()
    build_env["GOOS"] = target_os
    subprocess.check_call(['go', 'build', '-ldflags=-s -w', '-o', 'out/ytt-lint-%s' % target_os, './cmd/ytt-lint/'], env=build_env, shell=False, cwd=util.getrootdir())


def all() -> None:
    build(target_os="linux")
    build(target_os="darwin")
    build(target_os="windows")

if __name__ == "__main__":
    
    arg = ""
    if len(sys.argv) > 1:
        arg = sys.argv[1]
    
    if arg == "--all" or arg == "all" or arg == "-a":
        all()
    else:
        build()
