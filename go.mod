module github.com/phil9909/ytt-lint

go 1.13

require (
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/k14s/ytt v0.27.1
	github.com/onsi/gomega v1.9.0
	go.starlark.net v0.0.0-20190219202100-4eb76950c5f0
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab // indirect
)

replace go.starlark.net => github.com/k14s/starlark-go v0.0.0-20200402152745-409c85f3828d // ytt branch
