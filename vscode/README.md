# ytt-lint README

## Features

Lint your yaml-files and ytt-templates with ytt-lint

## Requirements

install ytt-lint

## Extension Settings

None

<!--

* `myExtension.enable`: enable/disable this extension
* `myExtension.thing`: set to `blah` to do something
-->

## Known Issues

Everything

## Release Notes

### 0.0.4 - Concourse

- support linting concourse pipelines
- improve "pull from k8s" to make sure extracted schemas contain `kind` and `apiVersion`

### 0.0.3 - Custom Libraries

- support 'load' of custom lib.yaml
- improve schema extraction to include multiple k8s versions (now supporting Deployment v1 out of the box)
- linting is now less pedantic about values used as strings.

### 0.0.2 - MVP

MVP release. IMHO it is already quite useful.

### 0.0.1 - Internal

Internal Only
