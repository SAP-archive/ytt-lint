package yttlint

import (
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestValidate(t *testing.T) {
	g := NewGomegaWithT(t)

	data, err := ioutil.ReadFile("../../examples/lint/ingress.yaml")

	if err != nil {
		t.Fatalf("Could not read test file %v", err)
	}

	errors := Lint(string(data), "test", "json")

	g.Expect(errors).To(ConsistOf(
		LinterError{
			Msg: ".metadata.name expected string got: integer",
			Pos: "test:13",
		},
		LinterError{
			Msg: ".metadata.namespace expected string got: integer",
			Pos: "test:11",
		},
		LinterError{
			Msg: ".spec.rules[0].http.paths[0].backend.servicePort expected int-or-string got a computed value. Tip: use str(...) or int(...) to convert to int or string",
			Pos: "test:26",
		},
		LinterError{
			Msg: ".spec.rules[0].http.paths[1].backend missing required entry serviceName",
			Pos: "test:28",
		},
	))
}

func TestValidateInvalidYAML(t *testing.T) {
	g := NewGomegaWithT(t)

	data, err := ioutil.ReadFile("../../examples/lint/invalid-yaml.yaml")

	if err != nil {
		t.Fatalf("Could not read test file %v", err)
	}

	errors := Lint(string(data), "test", "json")

	g.Expect(errors).To(ConsistOf(
		LinterError{
			Msg: "mapping values are not allowed in this context",
			Pos: "test:3",
		},
	))
}
