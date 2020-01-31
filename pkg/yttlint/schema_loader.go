package yttlint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const schemaDir = `/go/src/github.com/k14s/ytt/k8s-schema/`

var schemaCache map[string]map[string]interface{}

func loadSchema(kind string) (map[string]interface{}, error) {
	kind = strings.ToLower(kind)

	schema, ok := schemaCache[kind]
	if ok {
		return schema, nil
	}

	scheamFile, err := os.Open(os.Getenv("HOME") + "/" + schemaDir + "/" + kind + ".json")
	if err != nil {
		return nil, fmt.Errorf("could not open schema file: %v\nYou might need to run schema.sh first", err)
	}
	defer scheamFile.Close()

	byteValue, err := ioutil.ReadAll(scheamFile)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("could not read schema file: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("could not unmarshal schema file: %v", err)
	}

	if schemaCache == nil {
		schemaCache = make(map[string]map[string]interface{})
	}

	schemaCache[kind] = result
	return result, nil
}
