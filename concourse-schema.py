#!/usr/bin/env python3

import json
import urllib.request


in_name  =  "https://raw.githubusercontent.com/cappyzawa/concourse-pipeline-jsonschema/master/concourse_jsonschema.json"
out_name = "vscode/schema/builtin/concourse.json"

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
json.dump(schema, open(out_name, "w"), indent=2)
