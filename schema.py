#!/usr/bin/env python3

import json
import os.path
import os
import urllib.request

def findRefs(o):
    if "$ref" in o and isinstance(o["$ref"], str):
        return [o["$ref"]]
 
    if isinstance(o, dict):
        res = []
        for val in o.values():
            res.extend(findRefs(val))
        return res
    
    if isinstance(o, list):
        res = []
        for val in o:
            res.extend(findRefs(val))
        return res

    if isinstance(o, str):
        return []
    
    else:
        raise Exception("Unsupported type %s" % type(o).__name__)


def extraceSchema(file):
    schema = json.load(open(file))
    definitions = schema["definitions"]

    for (name, root) in definitions.items():
        if "x-kubernetes-group-version-kind" not in root:
            continue
        #if name != "io.k8s.api.core.v1.Pod":
        #    continue

        res = {
            name: root
        }

        refs = findRefs(res)

        while len(refs) > 0:
            ref = refs[0]
            if ref.startswith("#/definitions/"):
                ref = ref[14:]
            else:
                raise Exception("unsupported reference to %s" % ref)

            if ref not in res:
                res[ref] = definitions[ref]
                refs.extend(findRefs(definitions[ref]))
            refs = refs[1:]

        res = {
            "$ref": "#/definitions/%s" % name,
            "definitions": res,
        }
        
        for gvk in root["x-kubernetes-group-version-kind"]:
            target_dir = os.path.join("schema", "core" if gvk["group"] == "" else gvk["group"], gvk["version"])
            target = os.path.join(target_dir, gvk["kind"].lower() + ".json")
            print(target)
            os.makedirs(target_dir, exist_ok=True)
            json.dump(res, open(target, "w"))

urlTemplate = "https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json"
cacheTemplate = "./cache/k8s-%s-swagger.json"

os.makedirs("./cache", exist_ok=True)
for version in ["1.10.0", "1.11.0", "1.12.0", "1.13.0", "1.14.0", "1.15.0", "1.16.0", "1.17.0", "1.18.0"]:
    url = urlTemplate % version
    cache = cacheTemplate % version
    if not os.path.isfile(cache):
        print("Downloading swagger.json for %s from %s" % (version, url))
        urllib.request.urlretrieve(url, cache)
    print("Extracting schemas for %s" % version)
    extraceSchema(cache)
