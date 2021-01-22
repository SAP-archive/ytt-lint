package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/SAP/ytt-lint/pkg/format"
	"github.com/SAP/ytt-lint/pkg/pull"
	"github.com/SAP/ytt-lint/pkg/yttlint"
)

var linter yttlint.Linter
var file, rootFolder string
var excludeList []string

func main() {
	var pedantic, pullFromK8S, autoImport bool
	var pullKubeconfig, pullContext string
	flag.StringVar(&file, "f", "-", "File to validate")
	flag.StringVar(&rootFolder, "root", "", "Root folder for validation (defaults to directory containing target file)")
	flag.BoolVar(&pedantic, "p", false, "Use pedantic linting mode")
	flag.BoolVar(&autoImport, "autoimport", false, "Automatically import schema of every custom resource defintion found during linting")
	flag.BoolVar(&pullFromK8S, "pull-from-k8s", false, "Pull crd schemas from Kubernetes cluster")
	flag.StringVar(&pullKubeconfig, "kubeconfig", "", "path to kubeconfig (used only for --pull-from-k8s)")
	flag.StringVar(&pullContext, "context", "", "context inside kubeconfig (used only for --pull-from-k8s)")
	outputFormat := flag.String("o", "human", "Output format: either human or json")
	flag.Parse()

	if pullFromK8S {
		fmt.Println("Pulling from k8s...")
		err := pull.Pull(pullKubeconfig, pullContext)
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

	errors := []yttlint.LinterError{}

	stdin := false
	if strings.HasPrefix(file, "-:") {
		parts := strings.SplitN(file, ":", 2)
		file = parts[1]
		stdin = true
	} else if file == "-" {
		stdin = true
	}

	if file != "-" && isEntryFileExclude() {
		fmt.Fprintf(os.Stderr, "Warning '%s' is excluded. Won't lint anything\n", file)
		formatter.Format(os.Stdout, errors)
		os.Exit(0)
	}

	if stdin {
		errors = lintReader(os.Stdin, file, autoImport)
	} else {
		err := filepath.Walk(file, func(path string, info os.FileInfo, _ error) error {
			if isFileExcluded(path) {
				return filepath.SkipDir
			}

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
			fileErrors := lintReader(fp, path, autoImport)
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

func lintReader(in io.Reader, filename string, autoImport bool) []yttlint.LinterError {
	reader := bufio.NewReader(in)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	return linter.Lint(string(data), filename, autoImport)
}

func getRootFolder() string {
	if rootFolder == "" {
		stat, err := os.Stat(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if stat.IsDir() {
			rootFolder = file
		} else {
			rootFolder = filepath.Dir(file)
		}
	}
	return rootFolder
}

func isEntryFileExclude() bool {
	root := getRootFolder()
	f := file

	for {
		isExcluded := isFileExcluded(f)
		if isExcluded {
			return true
		}

		f = filepath.Dir(f)
		rel, err := filepath.Rel(root, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if rel == "." {
			return false
		}
	}

}

func isFileExcluded(filename string) bool {
	root := getRootFolder()

	if excludeList == nil {
		ignoreFile := path.Join(root, ".ytt-lint", "ignore")
		list, err := ioutil.ReadFile(ignoreFile)
		if err != nil {
			if os.IsNotExist(err) {
				excludeList = []string{}
				return false
			}
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		excludeList = []string{}
		excludeListUnfiltered := strings.Split(string(list), "\n")
		for _, item := range excludeListUnfiltered {
			item = strings.TrimSuffix(strings.TrimSpace(item), "/")
			if item == "" || strings.HasPrefix(item, "#") {
				continue
			}
			excludeList = append(excludeList, item)
		}
	}

	filename, err := filepath.Rel(root, filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if strings.HasPrefix(filename, "../") {
		fmt.Fprintln(os.Stderr, "file is outside root")
		os.Exit(1)
	}

	for _, exclude := range excludeList {
		matched, err := filepath.Match(exclude, filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if matched {
			return true
		}
	}
	return false
}
