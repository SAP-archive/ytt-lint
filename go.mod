module github.com/SAP/ytt-lint

go 1.13

require (
	github.com/adrg/strutil v0.2.3
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/k14s/ytt v0.28.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.8.1
	go.starlark.net v0.0.0-20190219202100-4eb76950c5f0
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	sigs.k8s.io/yaml v1.1.0
)

replace go.starlark.net => github.com/k14s/starlark-go v0.0.0-20200522161834-8a7b2030a110 // ytt branch
