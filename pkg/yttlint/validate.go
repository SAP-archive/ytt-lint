package yttlint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/k14s/ytt/pkg/cmd/core"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/template"
	tplcore "github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/k14s/ytt/pkg/yamltemplate"
	"github.com/k14s/ytt/pkg/yttlibrary"
	"go.starlark.net/starlark"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	_ "github.com/SAP/ytt-lint/pkg/librarywrapper" // inject into lib
	"github.com/SAP/ytt-lint/pkg/magic"
)

type myTemplateLoader struct {
	*workspace.TemplateLoader
	compiledTemplate *template.CompiledTemplate
	name             string
	api              yttlibrary.API
}

var _ template.CompiledTemplateLoader = myTemplateLoader{}

func (l myTemplateLoader) FindCompiledTemplate(module string) (*template.CompiledTemplate, error) {
	if module == l.name {
		return l.compiledTemplate, nil
	}
	return l.TemplateLoader.FindCompiledTemplate(module)
}

func (l myTemplateLoader) Load(
	thread *starlark.Thread, module string) (starlark.StringDict, error) {

	if strings.HasPrefix(module, "@ytt:") {
		if module == "@ytt:data" {
			return starlark.StringDict{
				"data": &magic.MagicType{},
			}, nil
		}
		res, err := l.api.FindModule(module[5:])
		if err == nil {
			return res, nil
		}
	}

	return l.TemplateLoader.Load(thread, module)
}

func (l myTemplateLoader) FilePaths(string) ([]string, error) {
	return nil, fmt.Errorf("unexpected call to FilePaths") // this should be handled by the injected magic-data-type
}

func (l myTemplateLoader) FileData(string) ([]byte, error) {
	return nil, fmt.Errorf("unexpected call to FileData") // this should be handled by the injected magic-data-type
}

func (l myTemplateLoader) LoadData(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return nil, fmt.Errorf("loadData is not supported")
}

func (l myTemplateLoader) ListData(
	thread *starlark.Thread, f *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	return nil, fmt.Errorf("listData is not supported")
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
	case float64:

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
func (l *Linter) Lint(data, filename string, autoImport bool) (errors []LinterError) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Recovered '%v' while linting %s \n", r, filename)
			errors = []LinterError{{
				Msg: fmt.Sprintf("could not lint because of an internal error: %v", r),
				Pos: fmt.Sprintf("%s:1", filename),
			}}
		}
	}()
	errors = l.lint(data, filename, autoImport)
	return
}

var helmChartRegex = regexp.MustCompile("{{")

func (l *Linter) lint(data, filename string, autoImport bool) []LinterError {
	helm := helmChartRegex.FindStringIndex(data)
	if helm != nil {
		firstHelmLine := strings.Count(data[:helm[0]], "\n") + 1
		return []LinterError{{
			Msg:  "This looks like helm syntax, skipping this file",
			Pos:  fmt.Sprintf("%s:%d", filename, firstHelmLine),
			Code: ErrorCodeHelm,
		}}
	}

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
	loader.TemplateLoader = workspace.NewTemplateLoader(workspace.NewEmptyDataValues(), []*workspace.DataValues{}, core.NewPlainUI(false), workspace.TemplateLoaderOpts{
		IgnoreUnknownComments: true,
	}, nil)
	var rootLib *workspace.Library
	loader.api, rootLib = newAPIandLib(filename, compiledTemplate.TplReplaceNode, loader)
	thread := &starlark.Thread{Name: "test", Load: loader.Load}

	thread.SetLocal("ytt.curr_library_key", rootLib)
	thread.SetLocal("ytt.root_library_key", rootLib)

	_, newVal, err := compiledTemplate.Eval(thread, loader)
	if err != nil {
		multiErr, ok := err.(template.CompiledTemplateMultiError)
		if ok {
			return mapMultierrorToLinterror(multiErr, filename)
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

	if autoImport {
		err = importCRDs(filename, newVal.(*yamlmeta.DocumentSet))
		if err != nil {
			panic(fmt.Errorf("autoimport failed: %w", err))
		}
	}

	for _, doc := range newVal.(*yamlmeta.DocumentSet).Items {
		gvk, item := extractKind(doc)
		var err error
		var schema *v1.JSONSchemaProps
		if gvk.kind != "" {
			schema, err = loadK8SSchema(gvk)
			if err != nil {
				errors = append(errors,
					appendLocationIfKnownf(item, "Error loading schema for kind %s: %v\n", gvk.kind, err.Error()))
				continue
			}
		} else if isConcoursePipeline(doc) {
			schema, err = loadConcourseSchema()
			if err != nil {
				panic(err)
			}
		} else {
			continue
			// TODO: print warning if not a trivial document
		}

		subSchema := convert(doc.Value)
		if subSchema.Description == "" && doc.Position.IsKnown() {
			doc.Position.SetFile(filename)
			subSchema.Description = doc.Position.AsString()
		}

		subErrors := l.isSubset(schema.Definitions, subSchema, schema, "")
		errors = append(errors, subErrors...)
	}

	return errors
}

func isConcoursePipeline(doc *yamlmeta.Document) bool {
	m, ok := doc.Value.(*yamlmeta.Map)
	if !ok {
		return false
	}

	for _, item := range m.Items {
		if item.Key == "jobs" {
			_, isArray := item.Value.(*yamlmeta.Array)
			return isArray
		}
	}

	return false
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

func convert(value interface{}) *v1.JSONSchemaProps {
	switch v := value.(type) {
	case *yamlmeta.Map:
		object := v1.JSONSchemaProps{
			Type:       "object",
			Properties: map[string]v1.JSONSchemaProps{},
		}

		for _, item := range v.Items {
			value := convert(item.Value)
			key := item.Key.(string)

			if strings.HasPrefix(key, "__ytt_lint_t_") || strings.HasPrefix(key, "__ytt_lint_f_") {
				key = key[13:]
			}

			if value.Description == "" && item.Position.IsKnown() {
				value.Description = item.Position.AsCompactString()
			}

			_, allreadExists := object.Properties[key]
			if allreadExists {
				oldValue := object.Properties[key]
				if len(oldValue.AnyOf) > 0 {
					oldValue.AnyOf = append(oldValue.AnyOf, *value)
				} else {
					value = &v1.JSONSchemaProps{
						AnyOf: []v1.JSONSchemaProps{oldValue, *value},
					}
				}
			}

			object.Properties[key] = *value
		}

		return &object

	case *yamlmeta.Array:
		items := make([]v1.JSONSchemaProps, 0, len(v.Items))
		for _, item := range v.Items {
			convertedItem := convert(item.Value)

			if convertedItem.Description == "" && item.Position.IsKnown() {
				convertedItem.Description = item.Position.AsCompactString()
			}

			items = append(items, *convertedItem)
		}

		return &v1.JSONSchemaProps{
			Type: "array",
			Items: &v1.JSONSchemaPropsOrArray{
				JSONSchemas: items,
			},
		}

	case []interface{}:
		items := make([]v1.JSONSchemaProps, 0, len(v))
		for _, item := range v {
			convertedItem := convert(item)
			items = append(items, *convertedItem)
		}
		return &v1.JSONSchemaProps{
			Type: "array",
			Items: &v1.JSONSchemaPropsOrArray{
				JSONSchemas: items,
			},
		}

	case string:
		encoded, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return &v1.JSONSchemaProps{
			Type: "string",
			Default: &v1.JSON{
				Raw: encoded,
			},
		}
	case int:
		encoded, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return &v1.JSONSchemaProps{
			Type: "integer",
			Default: &v1.JSON{
				Raw: encoded,
			},
		}
	case int64:
		encoded, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return &v1.JSONSchemaProps{
			Type: "integer",
			Default: &v1.JSON{
				Raw: encoded,
			},
		}
	case int32:
		encoded, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return &v1.JSONSchemaProps{
			Type: "integer",
			Default: &v1.JSON{
				Raw: encoded,
			},
		}
	case bool:
		encoded, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return &v1.JSONSchemaProps{
			Type: "boolean",
			Default: &v1.JSON{
				Raw: encoded,
			},
		}
	case *magic.MagicType:
		encoded, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return &v1.JSONSchemaProps{
			Type: "magic",
			Default: &v1.JSON{
				Raw: encoded,
			},
		}
	case nil:
		return &v1.JSONSchemaProps{
			Type: "null",
			Default: &v1.JSON{
				Raw: []byte("null"),
			}}

	default:
		return &v1.JSONSchemaProps{
			Type: "error",
			Default: &v1.JSON{
				Raw: []byte(fmt.Sprintf("convert(): unsupported type %T", value)),
			},
		}
	}

}

func (l *Linter) isSubset(defs v1.JSONSchemaDefinitions, subSchema, schema *v1.JSONSchemaProps, path string) []LinterError {
	errors := make([]LinterError, 0)

	if len(subSchema.AnyOf) > 0 {
		for _, item := range subSchema.AnyOf {
			subErrors := l.isSubset(defs, &item, schema, path)
			errors = append(errors, subErrors...)
		}
		return errors
	}

	if schema.Ref != nil {
		ref := *schema.Ref
		if strings.HasPrefix(ref, "#/definitions/") {
			deref := defs[ref[14:]]
			return l.isSubset(defs, subSchema, &deref, path)
		}
	}

	switch schema.Type {
	case "object":
		for key, prop := range schema.Properties {
			if subSchema.Type == "object" {

				subProp, ok := subSchema.Properties[key]
				if !ok {
					subPropT, okT := subSchema.Properties["__ytt_lint_t_"+key]
					if okT {
						subErrors := l.isSubset(defs, &subPropT, &prop, fmt.Sprintf("%s.%s", path, key))
						errors = append(errors, subErrors...)
					}
					subPropF, okF := subSchema.Properties["__ytt_lint_f_"+key]
					if okF {
						subErrors := l.isSubset(defs, &subPropF, &prop, fmt.Sprintf("%s.%s", path, key))
						errors = append(errors, subErrors...)
					}

					if !okF || !okT {
						for _, requiredKey := range schema.Required {
							if requiredKey == key {
								errors = append(errors, appendLocationIfKnownf(subSchema, "%s missing required entry %s", path, key))
								//break
							}
						}
					}
				} else {
					subErrors := l.isSubset(defs, &subProp, &prop, fmt.Sprintf("%s.%s", path, key))
					errors = append(errors, subErrors...)
				}
			} else {
				errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected object got: %s", path, subSchema.Type))
			}
		}

		var additionalPropertiesSchema *v1.JSONSchemaProps
		additionalPropertiesAllowAll := false
		if schema.AdditionalProperties != nil {
			additionalPropertiesSchema = schema.AdditionalProperties.Schema
			additionalPropertiesAllowAll = schema.AdditionalProperties.Allows && (additionalPropertiesSchema == nil)
		}
		if !additionalPropertiesAllowAll {
			for key, val := range subSchema.Properties {
				_, ok := schema.Properties[key]
				if !ok {
					if additionalPropertiesSchema != nil {
						subErrors := l.isSubset(defs, &val, additionalPropertiesSchema, fmt.Sprintf("%s.%s", path, key))
						errors = append(errors, subErrors...)
					} else {
						errors = append(errors, generateAdditionalPropertiesError(&val, path, key, schema.Properties))
					}
				}
			}
		}

	case "array":
		if subSchema.Type != "array" {
			if subSchema.Type == "magic" {
				if l.Pedantic {
					errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected array got a computed value`, path))
				}
			} else {
				errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected array got: %s", path, subSchema.Type))
			}
		} else {
			itemsSchema := schema.Items.Schema
			for i, item := range subSchema.Items.JSONSchemas {
				subErrors := l.isSubset(defs, &item, itemsSchema, fmt.Sprintf("%s[%d]", path, i))
				errors = append(errors, subErrors...)
			}
		}

	case "string":
		if subSchema.Type == "string" {
			if schema.Pattern != "" {
				pattern, err := regexp.Compile("^" + schema.Pattern + "$")
				if err != nil {
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s could not validate value, schema contains invalid pattern: %s", path, schema.Pattern))
				} else {
					if !pattern.MatchString(extractStringFromSchema(subSchema)) {
						errors = append(errors, appendLocationIfKnownf(subSchema, "%s invalid value. Expected to match pattern: %s", path, schema.Pattern))
					}
				}
			}
		} else {
			format := schema.Format

			if format == "int-or-string" {
				if subSchema.Type == "magic" {
					magic := extractMagicTypeFromSchema(subSchema)
					if l.Pedantic && !((magic.CouldBeString || magic.CouldBeInt) && !magic.CouldBeFloat) {
						errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected int-or-string got a computed value. Tip: use str(...) or int(...) to convert to int or string`, path))
					}
				} else if subSchema.Type != "integer" {
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected int-or-string got: %s", path, subSchema.Type))
				}
			} else {
				if subSchema.Type == "magic" {
					magic := extractMagicTypeFromSchema(subSchema)
					if l.Pedantic && !(magic.CouldBeString && !magic.CouldBeInt && !magic.CouldBeFloat) {
						errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected string got a computed value. Tip: use str(...) to convert to string`, path))
					}
				} else if path != ".metadata.creationTimestamp" { // https://github.com/kubernetes-sigs/controller-tools/issues/402
					errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected string got: %s", path, subSchema.Type))
				}
			}
		}

	case "integer":
		if subSchema.Type != "integer" {
			if subSchema.Type == "magic" {
				magic := extractMagicTypeFromSchema(subSchema)
				if l.Pedantic && !(magic.CouldBeInt && !magic.CouldBeString && !magic.CouldBeFloat) {
					errors = append(errors, appendLocationIfKnownf(subSchema, `%s expected integer got a computed value. Tip: use int(...) to convert to int`, path))
				}
			} else {
				errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected integer got: %s", path, subSchema.Type))
			}
		}
	case "boolean":
		if subSchema.Type != "boolean" {
			errors = append(errors, appendLocationIfKnownf(subSchema, "%s expected boolean got: %s", path, subSchema.Type))
		}

	case "":

	default:
		errors = append(errors, appendLocationIfKnownf(subSchema, " unsupported type %s", schema.Type))
	}

	return errors
}

func extractMagicTypeFromSchema(schema *v1.JSONSchemaProps) *magic.MagicType {
	magic := &magic.MagicType{}
	err := json.Unmarshal(schema.Default.Raw, magic)
	if err != nil {
		panic(err)
	}
	return magic
}

func extractStringFromSchema(schema *v1.JSONSchemaProps) string {
	var data = ""
	err := json.Unmarshal(schema.Default.Raw, &data)
	if err != nil {
		panic(err)
	}
	return data
}

func appendLocationIfKnownf(object interface{}, format string, a ...interface{}) LinterError {
	lintError := lintErrorf(format, a...)
	lintError.Pos = ""

	jsonP, ok := object.(*v1.JSONSchemaProps)
	if ok {
		lintError.Pos = jsonP.Description
	}

	mi, ok := object.(*yamlmeta.MapItem)
	if ok && mi.Position.IsKnown() {
		lintError.Pos = mi.Position.AsCompactString()
	}

	return lintError
}

type typoCandidate struct {
	similarity float64
	key        string
}
type typoCandidateList []typoCandidate

func (tcl typoCandidateList) Len() int {
	return len(tcl)
}

func (tcl typoCandidateList) Swap(i, j int) {
	tcl[i], tcl[j] = tcl[j], tcl[i]
}

func (s typoCandidateList) Less(i, j int) bool {
	return s[i].similarity < s[j].similarity
}

func generateAdditionalPropertiesError(val *v1.JSONSchemaProps, path string, key string, schema map[string]v1.JSONSchemaProps) LinterError {
	message := fmt.Sprintf("%s.%s additional properties are not permitted", path, key)

	alternatives := []string{}
	candidates := typoCandidateList{}

	for candidate := range schema {
		similarity := strutil.Similarity(strings.ToLower(key), strings.ToLower(candidate), metrics.NewHamming())
		if similarity < 0.25 {
			continue
		}
		candidates = append(candidates, typoCandidate{
			similarity: similarity,
			key:        candidate,
		})
	}

	sort.Sort(candidates)

	if len(candidates) > 5 {
		candidates = candidates[:5]
	}

	for _, candidate := range candidates {
		alternatives = append(alternatives, candidate.key)
	}

	if len(alternatives) != 0 {
		message = fmt.Sprintf("%s. Did you mean: %s?", message, strings.Join(alternatives, ", "))
	}

	return appendLocationIfKnownf(val, message)
}

func newAPIandLib(filename string, replaceNodeFunc tplcore.StarlarkFunc, loader yttlibrary.DataLoader) (yttlibrary.API, *workspace.Library) {
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
	library := workspace.NewLibraryModule(workspace.LibraryExecutionContext{
		Current: rootLib,
		Root:    rootLib,
	}, libraryExecutionFactory, []*workspace.DataValues{})
	libraryModule := library.AsModule()

	api := yttlibrary.NewAPI(replaceNodeFunc, yttlibrary.NewDataModule(&yamlmeta.Document{}, loader), libraryModule)

	return api, rootLib
}
