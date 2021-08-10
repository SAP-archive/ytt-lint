[![REUSE status](https://api.reuse.software/badge/github.com/SAP/ytt-lint)](https://api.reuse.software/info/github.com/SAP/ytt-lint)

# ytt-lint

ytt-lint is designed for ytt-templates to generate Kubernetes artifacts.

But it is also useful for:

* Other ytt-templates
* Plain Kubernetes yaml files.

Validating CRDs is supported! See "Working with CRDs"

## Installation

Currently I would not recommend to use ytt-lint from the terminal, as it has no stable interface yet.
If you still want to try it, got to the releases on this repo.

I recommend using the VSCode / VSCodium extension.
You can find it in the [VisualStudio Marketplace](https://marketplace.visualstudio.com/items?itemName=phil9909.ytt-lint) and in the [Open VSX Registry](https://open-vsx.org/extension/phil9909/ytt-lint).
Also, you can download it from the releases page and then install it using `code --install-extension /path/to/extension` or `codium --install-extension /path/to/extension`.


## Working with CRDs

This will only work, if the CRD offers a schema.

To pull the CRD schemas from the current kubernetes cluster:

* On VSCode or VSCodium run the `ytt-lint: Pull crd schemas from Kubernetes cluster`.
* or on terminal run `ytt-lint --pull-from-k8s` (or with the extension installed `~/.*code/extensions/phil9909.ytt-lint-*/bin/ytt-lint --pull-from-k8s`).

The schemas will then be stored locally. You might need to run this from time to time, if you update a controller or install a new one to your cluster.

## Excluding files

ytt-lint supports a git-like ignore file. To make use of it create a folder called ".ytt-lint" in your projects-root and put a file called "ignore" in there.

## Troubleshooting

### VSCode or VSCodium does not show any linter errors

Make sure the file is classified as a yaml file.
You can see that in the bottom right of the window.

### Error loading schema for kind ...

ytt-lint does not have schema information built-in for every Kubernetes artifact out there.

If the problem occurs on a CR see [Working with CRDs](#working-with-crds).

If you think the kind should be integrated into ytt-lint open an issue on the repo and we will discuss.

## Reporting issues

Open an issue on this repo. If possible include a ytt-template or plain-yaml-file which causes the problem.

## Acknowledgments

This project heavily depends on [ytt](https://github.com/k14s/ytt), which itself is based on the [Starlark in Go](https://github.com/google/starlark-go) project.
Kudos also to the [Kubernetes project](https://kubernetes.io/), which provide JSON Schemas for their resources, [cappyzawa](https://github.com/cappyzawa), who created a [JSON Schema for Concourse pipelines](https://github.com/cappyzawa/concourse-pipeline-jsonschema) and [JSON Schema Store](https://www.schemastore.org/) for providing the schema for kustomization.yaml.

## Licensing

Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file and all other files in this repository are licensed under the "Apache License, v 2.0" except as noted otherwise in the [LICENSE](/LICENSE) file.
