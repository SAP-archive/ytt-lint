package yttlint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const schemaDir = `/git/ytt-lint/schema/`

type kubernetesGVK struct {
	group, version, kind string
}

var schemaCache map[string]map[string]interface{}

func loadSchema(gvk kubernetesGVK) (map[string]interface{}, error) {
	gvk.kind = strings.ToLower(gvk.kind)
	key := path.Join(gvk.group, gvk.version, gvk.kind)

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

	var result map[string]interface{}
	found := false
	for _, schemaPath := range schemaPaths {
		schemaFileName := path.Join(schemaPath, "k8s", key+".json")
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
		return nil, fmt.Errorf("could not find schema file for %s/%s/%s in: %v", gvk.group, gvk.version, gvk.kind, schemaPaths)
	}

	if schemaCache == nil {
		schemaCache = make(map[string]map[string]interface{})
	}
	schemaCache[key] = result
	return result, nil
}
