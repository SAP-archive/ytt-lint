package yttlint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"go.starlark.net/starlark"
)

type myTemplateLoader struct {
	compiledTemplate *template.CompiledTemplate
	name             string
}

var _ template.CompiledTemplateLoader = myTemplateLoader{}

func (l myTemplateLoader) FindCompiledTemplate(module string) (*template.CompiledTemplate, error) {
	if module == l.name {
		return l.compiledTemplate, nil
	}
	return nil, fmt.Errorf("FindCompiledTemplate(%s) is not supported", module)

	//	return l.compiledTemplate, nil
}

func (l myTemplateLoader) Load(
	thread *starlark.Thread, module string) (starlark.StringDict, error) {

	//return nil, fmt.Errorf("Load(%s) is not supported", module)

	thread.SetLocal("data", &magicType{})
	//fmt.Printf("---\n%s\n+++\n", thread.CallStack().String())

	return starlark.StringDict{
		"data": &magicType{},
	}, nil
}

func (l myTemplateLoader) LoadData(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return nil, fmt.Errorf("LoadData is not supported")
}

func (l myTemplateLoader) ListData(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return nil, fmt.Errorf("ListData is not supported")
}

// Lint applies linting to a given ytt template
func Lint(data, filename string) (*yamlmeta.DocumentSet, *template.CompiledTemplate) {
	docSet, err := yamlmeta.NewDocumentSetFromBytes([]byte(data), yamlmeta.DocSetOpts{AssociatedName: filename})
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	compiledTemplate, err := yamltemplate.NewTemplate(filename, yamltemplate.TemplateOpts{}).Compile(docSet)
	if err != nil {
		fmt.Printf("NewTemplate: %s\n", err.Error())
		os.Exit(1)
	}

	//fmt.Printf("### ast:\n")
	//docSet.Print(os.Stdout)

	//fmt.Printf("### template:\n%s\n", compiledTemplate.DebugCodeAsString())

	loader := myTemplateLoader{compiledTemplate: compiledTemplate, name: filename}
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		fmt.Printf("Eval: %s\n", err.Error())
		os.Exit(1)

	}

	//fmt.Printf("### result ast:\n")
	//newVal.(*yamlmeta.DocumentSet).Print(os.Stdout)

	//combinedDocBytes, err := newVal.(*yamlmeta.DocumentSet).AsBytesWithPrinter(nil)
	//if err != nil {
	//	fmt.Printf(err.Error())
	//	os.Exit(1)
	//}

	//fmt.Printf("### result\n")
	//fmt.Printf("%s\n", combinedDocBytes)

	//schemaBytes, err := newVal.(*yamlmeta.DocumentSet).AsBytesWithPrinter(
	//	func(w io.Writer) yamlmeta.DocumentPrinter { return &schemaPrinter{buf: w} })
	//if err != nil {
	//	fmt.Printf(err.Error())
	//	os.Exit(1)
	//}

	//fmt.Printf("### schema\n")
	//fmt.Printf("%s\n", schemaBytes)

	schema, err := loadSchema()
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	doc := newVal.(*yamlmeta.DocumentSet).Items[0]
	subSchema := convert(doc.Value)

	errors := isSubset(subSchema, schema, "")
	if len(errors) == 0 {
		fmt.Println("No errors found")
	} else {
		for _, err := range errors {
			fmt.Printf("error: %v\n", err)
		}
	}

	return docSet, compiledTemplate
}

type schemaPrinter struct {
	buf io.Writer
}

var _ yamlmeta.DocumentPrinter = &schemaPrinter{}

func (p *schemaPrinter) Print(item *yamlmeta.Document) error {
	schema := convert(item.Value)
	asJSON, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	p.buf.Write([]byte(asJSON))
	return nil
}

func convert(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case *yamlmeta.Map:
		properties := make(map[string]interface{})
		for _, item := range v.Items {
			value := convert(item.Value)
			if _, ok := value["source"]; !ok && item.Position.IsKnown() {
				value["source"] = item.Position
			}
			properties[item.Key.(string)] = value
		}
		object := map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}

		return object
	case *yamlmeta.Array:
		items := make([]interface{}, 0, len(v.Items))
		for _, item := range v.Items {
			items = append(items, convert(item.Value))
		}
		object := map[string]interface{}{
			"type":  "array",
			"items": items,
		}

		return object
	case string:
		return map[string]interface{}{
			"type":  "string",
			"const": v,
		}
	case int:
		return map[string]interface{}{
			"type":  "integer",
			"const": v,
		}
	case *magicType:
		return map[string]interface{}{
			"type":  "magic",
			"magic": value,
		} // anything could be here

	default:
		return map[string]interface{}{
			"error": fmt.Sprintf("unsupported type %T", value),
		}
	}

}

func isSubset(subSchema, schema map[string]interface{}, path string) []error {
	errors := make([]error, 0)

	switch schema["type"] {
	case "object":
		prop, ok := schema["properties"]
		if !ok {
			prop = map[string]interface{}{}
		}
		properties, ok := prop.(map[string]interface{})
		if !ok {
			errors = append(errors, appendLocationIfKnownf(subSchema, "%s can't cast properties: %v", path, schema["properties"]))
		} else {
			for key, prop := range properties {
				subProps, ok := subSchema["properties"].(map[string]interface{})
				if !ok {
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s.%s can't cast subschema properties: %v", path, key, subSchema["properties"]))
				} else {
					subProp := subProps[key]
					if subProp == nil {
						//fmt.Printf("subProp(%v) == nil\n", key)
						// TODO: check required
					} else {
						subErrors := isSubset(subProp.(map[string]interface{}), prop.(map[string]interface{}), fmt.Sprintf("%s.%s", path, key))
						errors = append(errors, subErrors...)
					}
				}
			}
		}

	case "array":
		itemsSchema := schema["items"].(map[string]interface{})
		subItems, ok := subSchema["items"].([]interface{})
		if !ok {
			errors = append(errors, appendLocationIfKnownf(subSchema, "%s can't cast subSchema items: %v", path, subSchema["items"]))
		} else {
			for i, item := range subItems {
				//fmt.Println(item)
				subErrors := isSubset(item.(map[string]interface{}), itemsSchema, fmt.Sprintf("%s[%d]", path, i))
				errors = append(errors, subErrors...)
			}
		}
	case "string":
		if subSchema["type"] != "string" {
			if subSchema["type"] == "magic" {
				magic := subSchema["magic"].(*magicType)
				if !(magic.CouldBeString && !magic.CouldBeInt && !magic.CouldBeFloat) {
					errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected string got a computed value. Tip: use str(...) to convert to string`, path))
				}
			} else {
				errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected string got: %s", path, subSchema["type"]))
			}
		}

	default:
		errors = append(errors, appendLocationIfKnownf(subSchema, " unsupported type %s", schema["type"]))
	}

	return errors
}

func appendLocationIfKnownf(object map[string]interface{}, format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)

	if source, ok := object["source"]; ok && source.(*filepos.Position).IsKnown() {
		pos := source.(*filepos.Position)
		return fmt.Errorf("%s @ %s", msg, pos.AsString())
	}

	return errors.New(msg)
}
