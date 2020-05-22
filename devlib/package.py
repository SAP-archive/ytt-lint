#!/usr/bin/env python3

import subprocess
import platform
import os
import sys
import util

import zipfile
import os
import json
import shutil

def build(target_os:str = "native") -> None:
    if target_os == "native":
        target_os = platform.system().lower()
    print("Packing for %s" % target_os)

    oses = [target_os]
    if target_os == "bundle":
        oses = ["linux", "darwin", "windows"]

    packagejson = os.path.join(util.getextensiondir(), "package.json")
    extension_version = json.load(open(packagejson))["version"]

    dst_name = "out/ytt-lint-%s-%s.vsix" % (extension_version, target_os)
    src_name = "vscode/ytt-lint-%s.vsix" % extension_version

    shutil.copy2(src_name, dst_name)
    zip = zipfile.ZipFile(dst_name, 'a', compression=zipfile.ZIP_DEFLATED)
    for bundel_os in oses:
        zip.write("out/ytt-lint-%s" % bundel_os, "extension/bin/ytt-lint-%s" % bundel_os)
    zip.close()

def all() -> None:
    build(target_os="linux")
    build(target_os="darwin")
    build(target_os="windows")
    build(target_os="bundle")

if __name__ == "__main__":
    
    arg = ""
    if len(sys.argv) > 1:
        arg = sys.argv[1]
    
    if arg == "--all" or arg == "all" or arg == "-a":
        all()
    else:
        build()
