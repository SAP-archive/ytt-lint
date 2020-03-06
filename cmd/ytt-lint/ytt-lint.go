package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/phil9909/ytt-lint/pkg/yttlint"
)

func main() {
	var file string
	flag.StringVar(&file, "f", "-", "File to validate")
	outputFormat := flag.String("o", "human", "Output format: either human or json")
	flag.Parse()

	if *outputFormat != "json" && *outputFormat != "human" {
		fmt.Fprintf(os.Stderr, "Unsupported output format '%s' use json or human\n", *outputFormat)
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
			fmt.Fprintf(os.Stderr, "%s is a directory", file)
			os.Exit(1)
		}

		data, err = ioutil.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	}
	yttlint.Lint(string(data), file, *outputFormat)
}
