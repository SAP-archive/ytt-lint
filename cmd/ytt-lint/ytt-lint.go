package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/k14s/ytt/pkg/yttlint"
)

func main() {
	file := flag.String("f", "-", "File to validate")
	flag.Parse()

	var data []byte

	if *file == "-" {
		var err error
		reader := bufio.NewReader(os.Stdin)
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	} else {
		stat, err := os.Stat(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if stat.IsDir() {
			fmt.Fprintf(os.Stderr, "%d is a directory", file)
			os.Exit(1)
		}

		data, err = ioutil.ReadFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	}
	yttlint.Lint(string(data), *file)
}
