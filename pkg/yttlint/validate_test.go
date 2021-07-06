package yttlint

import (
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestValidate(t *testing.T) {
	type test struct {
		filename          string
		nonPedanticErrors []LinterError
		pedanticErrors    []LinterError
	}

	cases := []test{{
		filename: "../../examples/lint/ingress.yaml",
		nonPedanticErrors: []LinterError{{
			Msg: ".metadata.name expected string got: integer",
			Pos: "test:13",
		}, {
			Msg: ".metadata.namespace expected string got: integer",
			Pos: "test:11",
		}, {
			Msg: ".spec.rules[0].http.paths[1].backend.resource missing required entry kind",
			Pos: "test:30",
		}, {
			Msg:  `.spec.rules[0].http.paths[1].backend.resource.KinD additional properties are not permitted. Did you mean: kind?`,
			Pos:  "test:32",
			Code: "",
		}, {
			Msg:  `.spec.rules[0].http.paths[1].backend.resource.kynd additional properties are not permitted. Did you mean: kind?`,
			Pos:  "test:33",
			Code: "",
		}},
		pedanticErrors: []LinterError{{
			Msg: ".spec.rules[0].http.paths[0].backend.servicePort expected int-or-string got a computed value. Tip: use str(...) or int(...) to convert to int or string",
			Pos: "test:26",
		}},
	}, {
		filename: "../../examples/lint/len.yaml",
		nonPedanticErrors: []LinterError{{
			Msg: ".metadata.namespace expected string got: integer",
			Pos: "test:7",
		}},
		pedanticErrors: []LinterError{},
	}, {
		filename:          "../../examples/lint/base64.yaml",
		nonPedanticErrors: []LinterError{},
		pedanticErrors:    []LinterError{},
	}, {
		filename: "../../examples/lint/invalid-yaml.yaml",
		nonPedanticErrors: []LinterError{{
			Msg: "mapping values are not allowed in this context",
			Pos: "test:3",
		}},
		pedanticErrors: []LinterError{},
	}, {
		filename: "../../examples/lint/load-not-found.yaml",
		nonPedanticErrors: []LinterError{{
			// TODO: might remove the hint as it will confuse extension users.
			Msg: "cannot load file-not-found.yaml: Expected to find file 'file-not-found.yaml' (hint: only files included via -f flag are available)",
			Pos: "test:2",
		}},
		pedanticErrors: []LinterError{},
	}, {
		filename:          "../../examples/lint/array-parameter.yaml",
		nonPedanticErrors: []LinterError{},
		pedanticErrors:    []LinterError{},
	}, {
		filename:          "../../examples/lint/kubebuilder.yaml",
		nonPedanticErrors: []LinterError{},
		pedanticErrors:    []LinterError{},
	}, {
		filename: "../../examples/lint/empty-pod.yaml",
		nonPedanticErrors: []LinterError{{
			Msg: ".metadata.labels.label invalid value. Expected to match pattern: (([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?",
			Pos: "test:7",
		}, {
			Msg: ".spec.containers expected array got: null",
			Pos: "test:9",
		}},
		pedanticErrors: []LinterError{{
			Msg: ".spec.imagePullSecrets expected array got a computed value",
			Pos: "test:10",
		}},
	}, {
		filename: "../../examples/lint/concourse-caches.yaml",
		nonPedanticErrors: []LinterError{{
			Msg: ".jobs[0].plan[1].config.caches[0] expected object got: string",
			Pos: "test:15",
		}, {
			Msg: ".jobs[0].plan[1].config.run.args expected array got: string",
			Pos: "test:18",
		}},
		pedanticErrors: []LinterError{},
	}, {
		filename:          "../../examples/lint/array-access.yaml",
		nonPedanticErrors: []LinterError{},
		pedanticErrors:    []LinterError{
			// TODO: warn that it might not be a sliceable
		},
	}}

	for _, testCase := range cases {

		t.Run(testCase.filename, func(t *testing.T) {
			g := NewGomegaWithT(t)

			data, err := ioutil.ReadFile(testCase.filename)

			if err != nil {
				t.Fatalf("Could not read test file %v", err)
			}

			linter := &Linter{
				Pedantic: false,
			}
			errors := linter.Lint(string(data), "test", false)
			g.Expect(errors).To(ConsistOf(testCase.nonPedanticErrors))

			linter = &Linter{
				Pedantic: true,
			}
			errors = linter.Lint(string(data), "test", false)
			g.Expect(errors).To(ConsistOf(append(testCase.nonPedanticErrors, testCase.pedanticErrors...)))
		})

	}
}
