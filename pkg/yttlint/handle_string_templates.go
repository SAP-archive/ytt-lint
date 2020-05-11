package yttlint

import (
	"fmt"
	"strings"

	"github.com/k14s/ytt/pkg/yamlmeta"
)

// currently we ignore string templates
func injectStringTemplateHandling(val interface{}) {
	if val == nil {
		return
	}

	switch typedVal := val.(type) {
	case *yamlmeta.DocumentSet:
		injectStringTemplateHandling(&typedVal.Metas)
		for _, item := range typedVal.Items {
			injectStringTemplateHandling(item)
		}

	case *yamlmeta.Map:
		injectStringTemplateHandling(&typedVal.Metas)
		for _, item := range typedVal.Items {
			injectStringTemplateHandling(item)
		}
	case *yamlmeta.MapItem:
		injectStringTemplateHandling(&typedVal.Metas)
		injectStringTemplateHandling(typedVal.Key)
		injectStringTemplateHandling(typedVal.Value)

	case *yamlmeta.Array:
		injectStringTemplateHandling(&typedVal.Metas)
		for _, item := range typedVal.Items {
			injectStringTemplateHandling(item)
		}
	case *yamlmeta.ArrayItem:
		injectStringTemplateHandling(&typedVal.Metas)
		injectStringTemplateHandling(typedVal.Value)

	case *yamlmeta.Document:
		injectStringTemplateHandling(&typedVal.Metas)
		injectStringTemplateHandling(typedVal.Value)

	case *[]*yamlmeta.Meta:
		filtered := []*yamlmeta.Meta{}
		for _, item := range *typedVal {
			if !strings.Contains(item.Data, "yaml/text-templated-strings") {
				filtered = append(filtered, item)
			}
		}
		*typedVal = filtered

	case string:
	case int:
	case bool:
	case float64:

	default:
		panic(fmt.Sprintf("unsupported type hit injectStringTemplateHandling %T", typedVal))
	}
}
