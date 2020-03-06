package yttlint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	"github.com/phil9909/ytt-lint/pkg/librarywrapper"
	"github.com/phil9909/ytt-lint/pkg/magic"
	"go.starlark.net/starlark"
)

type myTemplateLoader struct {
	compiledTemplate *template.CompiledTemplate
	name             string
	api              yttlibrary.API
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

	if strings.HasPrefix(module, "@ytt:") {
		if module == "@ytt:data" {
			return starlark.StringDict{
				"data": &magic.MagicType{},
			}, nil
		}
		res, ok := l.api[module]
		if ok {
			return res, nil
		}
	}

	return nil, fmt.Errorf(`load("%s", ...) is not supported by ytt-lint`, module)
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

func isIf(meta *yamlmeta.Meta) bool {
	// TODO: ignore whitespace
	return strings.HasPrefix(meta.Data, "@ if ")
}

func isElse(meta *yamlmeta.Meta) bool {
	// TODO: ignore whitespace
	return strings.HasPrefix(meta.Data, "@ else:")
}

func isEnd(meta *yamlmeta.Meta) bool {
	// TODO: ignore whitespace
	return strings.HasPrefix(meta.Data, "@ end")
}

func injectIfHandling(val interface{}) {
	if val == nil {
		return
	}

	switch typedVal := val.(type) {
	case *yamlmeta.DocumentSet:
		//injectIfHandling(typedVal.Metas)
		for _, item := range typedVal.Items {
			injectIfHandling(item)
		}

	case *yamlmeta.Map:

		prefix := ""
		for _, item := range typedVal.Items {
			for _, meta := range item.Metas {
				if isIf(meta) {
					prefix = "__ytt_lint_t_"

				}
				if isElse(meta) {
					prefix = "__ytt_lint_f_"
					// FIXME: do proper filtering
					item.Metas = []*yamlmeta.Meta{}
				}
				if isEnd(meta) {
					prefix = ""
				}
			}
			item.Key = prefix + fmt.Sprint(item.Key)
		}

		for _, item := range typedVal.Items {
			injectIfHandling(item)
		}
	case *yamlmeta.MapItem:
		injectIfHandling(typedVal.Key)
		injectIfHandling(typedVal.Value)

	case *yamlmeta.Array:
		//injectIfHandling(typedVal.Metas)
		for _, item := range typedVal.Items {
			injectIfHandling(item)
		}
	case *yamlmeta.ArrayItem:
		//injectIfHandling(typedVal.Metas)
		injectIfHandling(typedVal.Value)

	case *yamlmeta.Document:
		//injectIfHandling(typedVal.Metas)
		injectIfHandling(typedVal.Value)

	case string:
	case int:
	case bool:

	default:
		panic(fmt.Sprintf("unsupported type hit injectIfHandling %T", typedVal))
	}
}

var (
	lineErrRegexp = regexp.MustCompile(`^yaml: line (?P<num>\d+): (?P<msg>.+)$`)
)

type Linter struct {
	Pedantic bool
}

// Lint applies linting to a given ytt template
func (l *Linter) Lint(data, filename string, outputFormat string) []LinterError {
	errors := l.lint(data, filename)

	switch outputFormat {
	case "json":
		jsonErrors, err := json.Marshal(errors)
		if err != nil {
			fmt.Printf("Eval: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Println(string(jsonErrors))

	case "human":
		if len(errors) == 0 {
			fmt.Println("No errors found")
		} else {
			for _, err := range errors {
				fmt.Printf("error: %s @ %s\n", err.Msg, err.Pos)
			}
		}
		fmt.Println()
	}

	return errors
}

func (l *Linter) lint(data, filename string) []LinterError {
	docSet, err := yamlmeta.NewDocumentSetFromBytes([]byte(data), yamlmeta.DocSetOpts{AssociatedName: filename})
	if err != nil {
		msg := err.Error()

		match := lineErrRegexp.FindStringSubmatch(msg)
		line, err := strconv.Atoi(match[1])
		if err != nil {
			panic(err)
		}
		msg = match[2]

		return []LinterError{{
			Msg: msg,
			Pos: fmt.Sprintf("%s:%d", filename, line),
		}}
	}

	//fmt.Printf("### ast:\n")
	//docSet.Print(os.Stdout)
	injectIfHandling(docSet)
	injectStringTemplateHandling(docSet)
	//docSet.Print(os.Stdout)

	compiledTemplate, err := yamltemplate.NewTemplate(filename, yamltemplate.TemplateOpts{
		IgnoreUnknownComments: true,
	}).Compile(docSet)
	if err != nil {
		fmt.Printf("NewTemplate: %s\n", err.Error())
		os.Exit(1)
	}

	//fmt.Printf("### template:\n%s\n", compiledTemplate.DebugCodeAsString())
	loader := myTemplateLoader{compiledTemplate: compiledTemplate, name: filename}
	loader.api = newAPI(filename, compiledTemplate.TplReplaceNode, loader)
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		multiErr, ok := err.(template.CompiledTemplateMultiError)
		if ok {
			return mapMultierrorToLinterror(multiErr)
		}
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

	errors := make([]LinterError, 0)

	for _, doc := range newVal.(*yamlmeta.DocumentSet).Items {
		gvk, item := extractKind(doc)
		if gvk.kind == "" {
			continue
			// TODO: print warning if not a trivial document
		}
		schema, err := loadSchema(gvk)
		if err != nil {
			errors = append(errors,
				appendLocationIfKnownf(item, "Error loading schema for kind %s: %v\n", gvk.kind, err.Error()))
			continue
		}

		subSchema := convert(doc.Value)
		if _, ok := subSchema["source"]; !ok && doc.Position.IsKnown() {
			doc.Position.SetFile(filename)
			subSchema["source"] = doc.Position
		}

		subErrors := l.isSubset(subSchema, schema, "")
		errors = append(errors, subErrors...)
	}

	return errors
}

func extractKind(doc *yamlmeta.Document) (kubernetesGVK, *yamlmeta.MapItem) {
	var gvk kubernetesGVK
	var loc *yamlmeta.MapItem

	m, ok := doc.Value.(*yamlmeta.Map)
	if !ok {
		return gvk, nil
	}
	for _, item := range m.Items {
		if item.Key == "kind" {
			gvk.kind = item.Value.(string)
			loc = item
		} else if item.Key == "apiVersion" {
			gv := strings.SplitN(item.Value.(string), "/", 2)
			if len(gv) == 1 {
				gvk.group = "core"
				gvk.version = gv[0]
			} else {
				gvk.group = gv[0]
				gvk.version = gv[1]
			}
		}
	}
	if gvk.group == "" || gvk.kind == "" || gvk.version == "" {
		gvk.group = ""
		gvk.kind = ""
		gvk.version = ""
	}
	return gvk, loc
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
			key := item.Key.(string)

			if strings.HasPrefix(key, "__ytt_lint_t_") || strings.HasPrefix(key, "__ytt_lint_f_") {
				key = key[13:]
			}

			if _, ok := value["source"]; !ok && item.Position.IsKnown() {
				value["source"] = item.Position
			}

			_, allreadExists := properties[key]
			if allreadExists {
				oldValue := properties[key].(map[string]interface{})
				anyOf, isAnyOf := oldValue["anyOf"]
				if isAnyOf {
					oldValue["anyOf"] = append(anyOf.([]interface{}), value)
				} else {
					value = map[string]interface{}{
						"anyOf": []interface{}{oldValue, value},
					}
				}
			}

			properties[key] = value
		}
		object := map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}

		return object
	case *yamlmeta.Array:
		items := make([]interface{}, 0, len(v.Items))
		for _, item := range v.Items {
			convertedItem := convert(item.Value)

			if _, ok := convertedItem["source"]; !ok && item.Position.IsKnown() {
				convertedItem["source"] = item.Position
			}

			items = append(items, convertedItem)
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
	case int32:
		return map[string]interface{}{
			"type":  "integer",
			"const": v,
		}
	case int64:
		return map[string]interface{}{
			"type":  "integer",
			"const": v,
		}
	case bool:
		return map[string]interface{}{
			"type":  "boolean",
			"const": v,
		}
	case *magic.MagicType:
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

func getAndCast(in map[string]interface{}, key string) (map[string]interface{}, int) {
	prop, ok := in[key]
	if !ok {
		return map[string]interface{}{}, 1
	}
	properties, ok := prop.(map[string]interface{})
	if !ok {
		return nil, 2
	}

	return properties, 0
}

func (l *Linter) isSubset(subSchema, schema map[string]interface{}, path string) []LinterError {
	errors := make([]LinterError, 0)

	anyOf, isAnyOf := subSchema["anyOf"]
	if isAnyOf {
		for _, item := range anyOf.([]interface{}) {
			subErrors := l.isSubset(item.(map[string]interface{}), schema, path)
			errors = append(errors, subErrors...)
		}
		return errors
	}

	switch schema["type"] {
	case "object":
		properties, code := getAndCast(schema, "properties")
		if code == 2 {
			errors = append(errors, appendLocationIfKnownf(subSchema, "%s can't cast properties: %v", path, schema["properties"]))
		} else {
			subProps, ok := subSchema["properties"].(map[string]interface{})
			if !ok {
				errors = append(errors, appendLocationIfKnownf(subSchema, "%s can't cast subschema properties: %v", path, subSchema["properties"]))
			} else {
				for key, prop := range properties {
					subProp := subProps[key]
					if subProp == nil {
						subPropT := subProps["__ytt_lint_t_"+key]
						if subPropT != nil {
							subErrors := l.isSubset(subPropT.(map[string]interface{}), prop.(map[string]interface{}), fmt.Sprintf("%s.%s", path, key))
							errors = append(errors, subErrors...)
						}
						subPropF := subProps["__ytt_lint_f_"+key]
						if subPropF != nil {
							subErrors := l.isSubset(subPropF.(map[string]interface{}), prop.(map[string]interface{}), fmt.Sprintf("%s.%s", path, key))
							errors = append(errors, subErrors...)
						}

						if subPropF == nil || subPropT == nil {
							r, ok := schema["required"]
							if ok {
								required := r.([]interface{})
								for _, requiredKey := range required {
									if requiredKey == key {
										errors = append(errors, appendLocationIfKnownf(subSchema, "%s missing required entry %s", path, key))
									}
								}
							}
						}
						//fmt.Printf("subProp(%v) == nil\n", key)
						// TODO: check required
					} else {
						subErrors := l.isSubset(subProp.(map[string]interface{}), prop.(map[string]interface{}), fmt.Sprintf("%s.%s", path, key))
						errors = append(errors, subErrors...)
					}
				}

				additionalProperties, code := getAndCast(schema, "additionalProperties")
				if code == 2 {
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s can't cast additionalProperties: %v", path, schema["additionalProperties"]))
				} else if code == 1 {
					for key, val := range subProps {
						_, ok := properties[key]
						if !ok {
							errors = append(errors, appendLocationIfKnownf(val, "%s.%s additional properties are not permitted", path, key))
						}
					}
				} else {
					for key, val := range subProps {
						_, ok := properties[key]
						if !ok {
							subErrors := l.isSubset(val.(map[string]interface{}), additionalProperties, fmt.Sprintf("%s.%s", path, key))
							errors = append(errors, subErrors...)
						}
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
				subErrors := l.isSubset(item.(map[string]interface{}), itemsSchema, fmt.Sprintf("%s[%d]", path, i))
				errors = append(errors, subErrors...)
			}
		}
	case "string":
		if subSchema["type"] != "string" {
			format, hasFormat := schema["format"]

			if hasFormat && format == "int-or-string" {
				if subSchema["type"] == "magic" {
					magic := subSchema["magic"].(*magic.MagicType)
					if l.Pedantic && !((magic.CouldBeString || magic.CouldBeInt) && !magic.CouldBeFloat) {
						errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected int-or-string got a computed value. Tip: use str(...) or int(...) to convert to int or string`, path))
					}
				} else if subSchema["type"] != "integer" {
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected int-or-string got: %s", path, subSchema["type"]))
				}
			} else {
				if subSchema["type"] == "magic" {
					magic := subSchema["magic"].(*magic.MagicType)
					if !(magic.CouldBeString && !magic.CouldBeInt && !magic.CouldBeFloat) {
						errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected string got a computed value. Tip: use str(...) to convert to string`, path))
					}
				} else {
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected string got: %s", path, subSchema["type"]))
				}
			}

		}
	case "integer":
		if subSchema["type"] != "integer" {
			if subSchema["type"] == "magic" {
				magic := subSchema["magic"].(*magic.MagicType)
				if l.Pedantic && !(magic.CouldBeInt && !magic.CouldBeString && !magic.CouldBeFloat) {
					errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected integer got a computed value. Tip: use int(...) to convert to int`, path))
				}
			} else {
				errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected integer got: %s", path, subSchema["type"]))
			}
		}
	case "boolean":
		if subSchema["type"] != "boolean" {
			errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected boolean got: %s", path, subSchema["type"]))
		}

	default:
		errors = append(errors, appendLocationIfKnownf(subSchema, " unsupported type %s", schema["type"]))
	}

	return errors
}

func appendLocationIfKnownf(object interface{}, format string, a ...interface{}) LinterError {

	lintError := lintErrorf(format, a...)

	m, ok := object.(map[string]interface{})
	lintError.Pos = ""

	if ok {
		if source, ok := m["source"]; ok && source.(*filepos.Position).IsKnown() {
			pos := source.(*filepos.Position)
			lintError.Pos = pos.AsCompactString()
		}
	}

	mi, ok := object.(*yamlmeta.MapItem)
	if ok && mi.Position.IsKnown() {
		lintError.Pos = mi.Position.AsCompactString()
	}

	return lintError
}

func newAPI(filename string, replaceNodeFunc tplcore.StarlarkFunc, loader template.CompiledTemplateLoader) yttlibrary.API {
	libraryExecutionFactory := workspace.NewLibraryExecutionFactory(core.NewPlainUI(false), workspace.TemplateLoaderOpts{
		IgnoreUnknownComments: true,
	})

	inputFiles, err := files.NewSortedFilesFromPaths([]string{filepath.Dir(filename)}, files.SymlinkAllowOpts{
		AllowAll: true,
	})
	if err != nil {
		os.Exit(1)
	}
	rootLib := workspace.NewRootLibrary(inputFiles)
	libraryModule := workspace.NewLibraryModule(workspace.LibraryExecutionContext{
		Current: rootLib,
		Root:    rootLib,
	}, libraryExecutionFactory).AsModule()

	api := yttlibrary.NewAPI(replaceNodeFunc, &yamlmeta.Document{}, loader, libraryModule)
	api["@ytt:base64"] = librarywrapper.Base64APIWrapper

	return api
}
