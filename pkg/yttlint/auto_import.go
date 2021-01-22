package yttlint

import (
	"fmt"
	"os"

	"github.com/SAP/ytt-lint/pkg/importer"
	"github.com/k14s/ytt/pkg/yamlfmt"
	"github.com/k14s/ytt/pkg/yamlmeta"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"sigs.k8s.io/yaml"
)

func importCRDs(filename string, docs *yamlmeta.DocumentSet) error {
	printer := yamlfmt.NewPrinter(nil)
	importer, err := importer.NewImporter()
	if err != nil {
		return err
	}

	for _, doc := range docs.Items {
		gvk, _ := extractKind(doc)

		if gvk.group != "apiextensions.k8s.io" {
			continue
		}
		if gvk.kind != "CustomResourceDefinition" {
			continue
		}

		asYaml := printer.PrintStr(&yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{doc},
		})

		switch gvk.version {
		case "v1":
			crd := v1.CustomResourceDefinition{}
			yaml.Unmarshal([]byte(asYaml), &crd)
			importer.ImportV1(crd)

		case "v1beta1":
			crd := v1beta1.CustomResourceDefinition{}
			yaml.Unmarshal([]byte(asYaml), &crd)
			importer.ImportV1Beta1(crd)

		default:
			fmt.Fprintf(os.Stderr, "autoimport warning: found CustomResourceDefinition of unsupported version %s in file %s (currently supported is v1 and v1beta1)\n", gvk.version, filename)
		}

	}
	return nil
}
