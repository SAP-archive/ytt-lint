package yttlint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const schemaDir = `/git/ytt-lint/vscode/schema/`

type kubernetesGVK struct {
	group, version, kind string
}

var schemaCache map[string]*v1.JSONSchemaProps

func loadK8SSchema(gvk kubernetesGVK) (*v1.JSONSchemaProps, error) {
	gvk.kind = strings.ToLower(gvk.kind)
	key := path.Join("k8s", gvk.group, gvk.version, gvk.kind)
	return loadSchema(key)
}
func loadConcourseSchema() (*v1.JSONSchemaProps, error) {
	return loadSchema(path.Join("builtin", "concourse"))
}

func loadSchema(key string) (*v1.JSONSchemaProps, error) {
	schema, ok := schemaCache[key]
	if ok {
		return schema, nil
	}

	schemaPaths := []string{
		path.Join(os.Getenv("HOME"), ".ytt-lint", "schema"),
	}
	if schemaPath, ok := os.LookupEnv("YTT_LINT_SCHEMA_PATH"); ok {
		schemaPaths = append(schemaPaths, strings.Split(schemaPath, ":")...)
	} else {
		schemaPaths = append(schemaPaths, path.Join(os.Getenv("HOME"), schemaDir))
	}

	result := &v1.JSONSchemaProps{}
	found := false
	for _, schemaPath := range schemaPaths {
		schemaFileName := path.Join(schemaPath, key+".json")
		scheamFile, err := os.Open(schemaFileName)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("could not open schema file: %v", err)
		}
		defer scheamFile.Close()

		byteValue, err := ioutil.ReadAll(scheamFile)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("could not read schema file: %v", err)
		}

		err = json.Unmarshal([]byte(byteValue), &result)

		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("could not unmarshal schema file: %v", err)
		}

		found = true
		break
	}

	if !found {
		return nil, fmt.Errorf("could not find schema file %s.json in: %v", key, schemaPaths)
	}

	if schemaCache == nil {
		schemaCache = make(map[string]*v1.JSONSchemaProps)
	}
	schemaCache[key] = result
	return result, nil
}
