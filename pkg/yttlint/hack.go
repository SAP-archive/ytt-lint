package yttlint

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/template"
)

func mapMultierrorToLinterror(multiErr template.CompiledTemplateMultiError, rootFile string) []LinterError {
	errors := make([]LinterError, 0)
	errs := reflect.ValueOf(multiErr).FieldByName("errs")
	n := errs.Len()
	for i := 0; i < n; i++ {
		msg := errs.Index(i).FieldByName("Msg").String()
		positions := errs.Index(i).FieldByName("Positions")
		m := positions.Len()
		for j := 0; j < m; j++ {
			var sourceLine reflect.Value

			position := positions.Index(j)

			sourceLine = getSourceLine(position.FieldByName("TemplateLine"))

			if !sourceLine.IsValid() || sourceLine.IsNil() {
				sourceLine = getSourceLine(position.FieldByName("BeforeTemplateLine"))
			}
			if !sourceLine.IsValid() || sourceLine.IsNil() {
				continue
			}

			pos := sourceLine.Elem().FieldByName("Position").Elem()
			errors = append(errors, LinterError{
				Msg: msg,
				Pos: fmt.Sprintf("%s:%d", pos.FieldByName("file").String(), pos.FieldByName("line").Elem().Int()),
			})
		}
		if m == 0 {
			errors = append(errors, LinterError{
				Msg: msg,
				Pos: fmt.Sprintf("%s:%d", rootFile, 1),
			})
		}
	}
	return errors
}

func getSourceLine(templateLine reflect.Value) reflect.Value {
	if !templateLine.IsNil() {
		return templateLine.Elem().FieldByName("SourceLine")
	}
	return reflect.Value{}
}
