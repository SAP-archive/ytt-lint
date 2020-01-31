package yttlint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const schema = `/go/src/github.com/k14s/ytt/k8s-schema/ingress.json`

func loadSchema() (map[string]interface{}, error) {

	scheamFile, err := os.Open(os.Getenv("HOME") + "/" + schema)
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

	return result, nil
}
