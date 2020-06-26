package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/ytt-lint/pkg/format"
	"github.com/SAP/ytt-lint/pkg/pull"
	"github.com/SAP/ytt-lint/pkg/yttlint"
)

var linter yttlint.Linter

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

	linter = yttlint.Linter{
		Pedantic: pedantic,
	}

	var in io.Reader
	errors := []yttlint.LinterError{}

	if file == "-" || strings.HasPrefix(file, "-:") {
		in = os.Stdin
		parts := strings.SplitN(file, ":", 2)
		if len(parts) == 2 {
			file = parts[1]
		}

		errors = lintReader(in, file)
	} else {
		err := filepath.Walk(file, func(path string, info os.FileInfo, _ error) error {
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") && file != path {
				return nil
			}

			fp, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			defer fp.Close()
			fileErrors := lintReader(fp, path)
			errors = append(errors, fileErrors...)

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	}

	formatter.Format(os.Stdout, errors)
}

func lintReader(in io.Reader, filename string) []yttlint.LinterError {
	reader := bufio.NewReader(in)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	return linter.Lint(string(data), filename)
}
