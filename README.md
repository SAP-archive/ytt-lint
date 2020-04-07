
# ytt-lint

ytt-lint is designed for ytt-templates to generate Kubernetes artifacts.

But it is also useful for:

* Other ytt-templates
* Plain Kubernetes yaml files.

Validating CRDs is supported! See "Working with CRDs"

## Installation

Currently I would not recommend to use ytt-lint from the terminal, as it has no stable interface yet.
If you still want to try it, got to the releases on this repo.

I recommend using the VSCode / VSCodium extension. You can download it from the releases page and then install it using `code --install-extension /path/to/extension` or `codium --install-extension /path/to/extension`.

Notice: Currently the extension contains a ytt-executable. Because of that the extension file is operating system specific.
This is because the ytt interface is not yet stable and it make sense to bundle them. This might change in the future.

## Working with CRDs

This will only work, if the CRD offers a schema.

To pull the CRD schemas from the current kubernetes cluster:

* On VSCode or VSCodium run the `ytt-lint: Pull crd schemas from Kubernetes cluster`.
* or on terminal run `ytt-lint --pull-from-k8s` (or with the extension installed `~/.*code/extensions/phil9909.ytt-lint-*/bin/ytt-lint --pull-from-k8s`).

The schemas will then be stored locally. You might need to run this from time to time, if you update a controller or install a new one to your cluster.

## Troubleshooting

### VSCode or VSCodium does not show any linter errors

Make sure the file is classified as a yaml file.
You can see that in the bottom right of the window.

### Error loading schema for kind ...

ytt-lint does not have schema information built-in for every Kubernetes artifact out there.

If the problem occurs on a CR see (Working with CRDs)[#working-with-crds].

If you think the kind should be integrated into ytt-lint open an issue on the repo and we will discuss.

## Reporting issues

Open an issue on this repo. If possible include a ytt-template or plain-yaml-file which causes the problem.

# Acknowledgments

This project heavily depends on (ytt)[https://github.com/k14s/ytt], which itself is based on the (Starlark in Go)[https://github.com/google/starlark-go] project.
Kudos also to the (Kubernetes project)[https://kubernetes.io/], which provide JSON Schemas for their resources, and (cappyzawa)[https://github.com/cappyzawa], who created a (JSON Schema for Concourse pipelines)[https://github.com/cappyzawa/concourse-pipeline-jsonschema].
