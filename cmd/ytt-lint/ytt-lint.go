package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/ytt-lint/pkg/format"
	"github.com/SAP/ytt-lint/pkg/pull"
	"github.com/SAP/ytt-lint/pkg/yttlint"
)

func main() {
	var file string
	var pedantic, pullFromK8S bool
	flag.StringVar(&file, "f", "-", "File to validate")
	flag.BoolVar(&pedantic, "p", false, "Use pedantic linting mode")
	flag.BoolVar(&pullFromK8S, "pull-from-k8s", false, "Pull crd schemas from Kubernetes cluster")
	outputFormat := flag.String("o", "human", "Output format: either human or json")
	flag.Parse()

	if pullFromK8S {
		fmt.Println("Pulling from k8s...")
		err := pull.Pull()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	formatter, err := format.GetFormatter(format.Format(*outputFormat))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}

	var data []byte

	if file == "-" || strings.HasPrefix(file, "-:") {
		var err error
		reader := bufio.NewReader(os.Stdin)
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		parts := strings.SplitN(file, ":", 2)
		if len(parts) == 2 {
			file = parts[1]
		}

	} else {
		stat, err := os.Stat(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if stat.IsDir() {
			errors := []yttlint.LinterError{}
			filepath.Walk(file, func(path string, info os.FileInfo, _ error) error {
				if info.IsDir() {
					return nil
				}
				if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
					return nil
				}

				data, err = ioutil.ReadFile(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					os.Exit(1)
				}
				linter := yttlint.Linter{
					Pedantic: pedantic,
				}
				fileErrors := linter.Lint(string(data), path)
				errors = append(errors, fileErrors...)

				return nil
			})
			formatter.Format(os.Stdout, errors)
			os.Exit(0)
		} else {
			data, err = ioutil.ReadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}

	}
	linter := yttlint.Linter{
		Pedantic: pedantic,
	}

	errors := linter.Lint(string(data), file)
	formatter.Format(os.Stdout, errors)
}
