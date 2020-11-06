#!/usr/bin/env python3

import subprocess
import json
import os.path
import sys
import hashlib

raw_mods = subprocess.check_output(["go", "list", "--json", "-m", "all"]).decode('UTF-8')
mods = [json.loads(x+"}") for x in raw_mods.split("\n}\n") if x.strip() != ""]

placeholdersInLicenses = {
    "BSD-2-Clause": {
        "<year>": "year",
        "<owner>": "licensor"
    },
    "BSD-3-Clause": {
        "<year>": "year",
        "<owner>": "licensor"
    },
    "MIT": {
        "<year>": "year",
        "<copyright holders>": "licensor"
    },
    "Apache-2.0": {},
    "ISC": {
        "<YEAR>": "year",
        "<OWNER>": "licensor"
    },
}

shaLookup = {}
if os.path.isfile('.notice/3rd-party-lookup.json'):
    with open('.notice/3rd-party-lookup.json') as file:
        shaLookup = json.load(file)

def saveLookup(sha, key, value):
    if sha not in shaLookup:
        shaLookup[sha] = {}
    shaLookup[sha][key] = value

    with open('.notice/3rd-party-lookup.json', 'w') as file:
        json.dump(shaLookup, file, indent=4)

def lookup(sha, key):
    if sha not in shaLookup:
        return None
    if key not in shaLookup[sha]:
        return None
    return shaLookup[sha][key]

def sha256(file):
    hash = hashlib.sha256()
    with open(file, 'r') as file:
        hash.update(file.read().encode('UTF-8'))
    return hash.hexdigest()


def detect(basedir, key, suffixes, fallback):
    for suffix in suffixes:
        file = basedir + "/" + suffix
        if os.path.isfile(file):
            sha = sha256(file)
            name = lookup(sha, key)
            if name is None:
                with open(file, 'r') as file:
                    print(file.read())
                name = input("What %s is this (%s): " % (key, file.name))
                saveLookup(sha, key, name)
            if name != "":
                return name

    name = lookup(fallback, key)
    if name is None:
        name = input("What %s is this '%s': " % (key, fallback))
        saveLookup(fallback, key, name)
    if name != "":
        return name
    
    print("Could not detect %s for gomod %s" % (key, mod))
    sys.exit(1)

licenses = [{
    "license": "MIT",
    "component": "Visual Studio Code extension generator 1.2.11",
    "website": "https://github.com/Microsoft/vscode-generator-code",
    "licensor": "Microsoft Corporation",
    "year": "unspecified",
}, {
    "license": "MIT",
    "component": "The Visual Studio Code Extension Manager",
    "licensor": "Microsoft Corporation",
    "website": "https://github.com/microsoft/vscode-vsce",
    "year": "unspecified",
}, {
    "license": "Apache-2.0",
    "component": "Kubernetes OpenAPI Spec",
    "licensor": "The Kubernetes Authors",
    "website": "https://github.com/kubernetes/kubernetes/blob/master/api/openapi-spec/swagger.json",
    "year": "unspecified",
}, {
    "license": "Apache-2.0",
    "component": "concourse-pipeline-jsonschema",
    "licensor": "Shu Kutsuzawa cappyzawa@yahoo.ne.jp",
    "website": "https://github.com/cappyzawa/concourse-pipeline-jsonschema",
    "year": "2020",
}, {
    "license": "Apache-2.0",
    "component": "kustomization.yaml from \"JSON Schema Store\"",
    "licensor": "Mads Kristensen",
    "website": "https://www.schemastore.org",
    "year": "2015-current",
}]

for mod in mods:
    if ("Main" in mod and mod["Main"]) or ("Indirect" in mod and mod["Indirect"]):
        continue
    license = detect(mod["Dir"], "license", ["LICENSE", "License"], mod["Path"])
    licensor = detect(mod["Dir"], "licensor", ["NOTICE", "LICENSE", "License"], mod["Path"])
    year = detect(mod["Dir"], "year", ["NOTICE", "LICENSE", "License"], mod["Path"])
    website = detect(mod["Dir"], "website", [], mod["Path"])
    #print(mod, license)
    licenses.append({
        "license": license,
        "licensor": licensor,
        "component": "%s %s" % (mod["Path"], mod["Version"]),
        "website": website,
        "year": year,
    })
    print(mod)


dependencies = []
with open("vscode/package.json") as file:
    package = json.load(file)
    if "dependencies" in package:
        dependencies.extend(package["dependencies"])
    if "devDependencies" in package:
        dependencies.extend(package["devDependencies"])
    #print(dependencies)

for dep in dependencies:
    with open("vscode/node_modules/" + dep + "/package.json") as file:
        package = json.load(file)
    
    author = ""
    if "author" in package and "name" in package["author"]:
        author = package["author"]["name"]
    else:
        author = detect("vscode/node_modules/" + dep, "licensor", ["NOTICE", "LICENSE", "License"], dep)

    year = detect("vscode/node_modules/" + dep, "year", ["NOTICE", "LICENSE", "License"], dep)
    
    licenses.append({
        "license": package["license"],
        "component": "%s %s" % (package["name"], package["version"]),
        "website": package["homepage"],
        "licensor": author,
        "year": year,
    })

with open("THIRD-PARTY-NOTICES.txt", "w") as target:
    print("ytt lint uses third-party software or other resources that may be distributed under licenses different than ytt lint software.", file=target)
    print("", file=target)
    print("In the event that we overlooked to list required notice, please bring this to our attention by contacting us via this email: opensource@sap.com", file=target)
    print("", file=target)
    print("", file=target)
    print("", file=target)
    print("------------------------------------------------------------------------------------------", file=target)
    print("", file=target)
    print("", file=target)
    print("Components:", file=target)

    for license in sorted(licenses, key=lambda  x: x["component"].lower()):
        print("", file=target)
        print("Component:", license["component"], file=target)
        print("Licensor:",  license["licensor"], file=target)
        print("Website:",   license["website"], file=target)
        print("License:",   license["license"], file=target)
        placeholders = placeholdersInLicenses[license["license"]]
        for placeholder in placeholders:
            lookupKey = placeholders[placeholder]
            print("%s = %s" % (placeholder, license[lookupKey]), file=target)

    print(licenses)
    for license in sorted(set([x["license"] for x in licenses])):
        #print(license)
        cache = "./.notice/license-%s" % license
        if not os.path.isfile(cache):
            print("License text for %s not found in %s. Try looking at https://github.com/spdx/license-list-data/tree/master/text or https://opensource.org/licenses/alphabetical" % (license, cache))
            sys.exit(1)
        print("\n\n" + (30*"-") + " " + license + " " + (30*"-") + "\n\n", file=target)
        with open(cache, 'r') as file:
            print(file.read(), file=target)
