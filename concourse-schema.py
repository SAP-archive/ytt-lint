#!/usr/bin/env python3

import json
import os
import os.path
import urllib.request
import devlib.util


in_name    = "https://raw.githubusercontent.com/cappyzawa/concourse-pipeline-jsonschema/0961a7b3d34fa3c6af28e0b3e93ef25fcf10c417/concourse_jsonschema.json"
target_dir = os.path.join(devlib.util.getextensiondir(), "schema", "builtin")
out_name   = os.path.join(target_dir, "concourse.json")

def adapt(data):
    if isinstance(data, dict):
        for k, v in data.items():
            data[k] = adapt(v)
        if "patternProperties" in data:
            data["additionalProperties"] = data["patternProperties"][".*"]
            del data["patternProperties"]
            if "additionalProperties" in data["additionalProperties"]:
                data["additionalProperties"] = True
    if isinstance(data, list):
        for k, v in enumerate(data):
            data[k] = adapt(v)
    return data

filedata = urllib.request.urlopen(in_name)
schema = json.load(filedata)
schema = adapt(schema)
schema["definitions"]["Config"]["additionalProperties"] = True
os.makedirs(target_dir, exist_ok=True)
json.dump(schema, open(out_name, "w"), indent=2)
