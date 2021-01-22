package yttlint

import (
	"fmt"
	"os"

	"github.com/SAP/ytt-lint/pkg/importer"
	"github.com/k14s/ytt/pkg/yamlfmt"
	"github.com/k14s/ytt/pkg/yamlmeta"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
		if gvk.version != "v1" {
			fmt.Fprintf(os.Stderr, "autoimport warning: found CustomResourceDefinition of unsupported version %s in file %s (currently supported is v1)\n", gvk.version, filename)
			continue
		}

		asYaml := printer.PrintStr(&yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{doc},
		})

		crd := apiextensionsv1.CustomResourceDefinition{}
		yaml.Unmarshal([]byte(asYaml), &crd)
		importer.ImportV1(crd)
	}
	return nil
}
