#!/usr/bin/env python3

import json
import os.path
import os
import urllib.request
import devlib.util

def findRefs(o):
    if isinstance(o, bool):
        return []

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

label_regex = r'(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?'

def extraceSchema(file):
    schema = json.load(open(file))
    definitions = schema["definitions"]

    definitions["io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"]["properties"]["labels"]["additionalProperties"]["pattern"] = label_regex

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
            target_dir = os.path.join(devlib.util.getextensiondir(), "schema", "k8s", "core" if gvk["group"] == "" else gvk["group"], gvk["version"])
            target = os.path.join(target_dir, gvk["kind"].lower() + ".json")
            print(target)
            os.makedirs(target_dir, exist_ok=True)
            json.dump(res, open(target, "w"))

urlTemplate = "https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json"
cacheTemplate = "./cache/k8s-%s-swagger.json"

swagger_files = []

os.makedirs("./cache", exist_ok=True)
for version in ["1.10.0", "1.11.0", "1.12.0", "1.13.0", "1.14.0", "1.15.0", "1.16.0", "1.17.0", "1.18.0", "1.19.0"]:
    swagger_files.append({
        "url": urlTemplate % version,
        "cache": cacheTemplate % version,
        "name": f"kubernetes@{version}"
    })

for swagger_file in swagger_files:
    if not os.path.isfile(swagger_file["cache"]):
        print("Downloading swagger.json for %s from %s" % (version, swagger_file["url"]))
        urllib.request.urlretrieve(swagger_file["url"], swagger_file["cache"])
    print("Extracting schemas for %s" % swagger_file["name"])
    extraceSchema(swagger_file["cache"])


def add_kustomize():
    target_dir = os.path.join(devlib.util.getextensiondir(), "schema", "k8s", "kustomize.config.k8s.io", "v1beta1")
    target = os.path.join(target_dir, "kustomization.json")

    print("adding support for kustomization file to", target)
    os.makedirs(target_dir, exist_ok=True)
    urllib.request.urlretrieve("https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/kustomization.json", target)

add_kustomize()
