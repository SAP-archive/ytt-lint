package format

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/SAP/ytt-lint/pkg/yttlint"
	"github.com/pkg/errors"
)

// Formatter turns arrays into a string (and prints them to a writer)
type Formatter interface {
	Format(writer io.Writer, errors []yttlint.LinterError) error
}

// Format name of a output format
type Format string

const (
	// FormatJSON constant for json format
	FormatJSON = Format("json")
	// FormatHuman constant for human readable format
	FormatHuman = Format("human")
)

// GetFormatter returns a formatter for a given string
func GetFormatter(format Format) (Formatter, error) {
	switch format {
	case FormatJSON:
		return &jsonFormatter{}, nil
	case FormatHuman:
		return &humanFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format '%s' use json or human", string(format))
	}
}

type jsonFormatter struct{}
type humanFormatter struct{}

func (*jsonFormatter) Format(writer io.Writer, lintErrors []yttlint.LinterError) error {
	jsonErrors, err := json.Marshal(lintErrors)
	if err != nil {
		return errors.Wrap(err, "could not marshal")
	}
	_, err = fmt.Fprintln(writer, string(jsonErrors))
	return errors.Wrap(err, "could not write")
}

func (*humanFormatter) Format(writer io.Writer, lintErrors []yttlint.LinterError) error {
	if len(lintErrors) == 0 {
		fmt.Fprintln(writer, "No errors found")
	} else {
		for _, err := range lintErrors {
			fmt.Fprintf(writer, "error: %s @ %s\n", err.Msg, err.Pos)
		}
	}
	_, err := fmt.Fprintln(writer)
	return errors.Wrap(err, "could not write")
}
