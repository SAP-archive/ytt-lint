module github.com/phil9909/ytt-lint

go 1.13

require (
	github.com/k14s/ytt v0.26.0
	github.com/onsi/gomega v1.9.0
	go.starlark.net v0.0.0-20190219202100-4eb76950c5f0
)

replace go.starlark.net => github.com/k14s/starlark-go v0.0.0-20200207164905-fd8842955e4e // ytt branch
