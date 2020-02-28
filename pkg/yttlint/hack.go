package yttlint

import (
	"fmt"
	"reflect"

	"github.com/k14s/ytt/pkg/template"
)

func mapMultierrorToLinterror(multiErr template.CompiledTemplateMultiError) []LinterError {
	errors := make([]LinterError, 0)
	errs := reflect.ValueOf(multiErr).FieldByName("errs")
	n := errs.Len()
	for i := 0; i < n; i++ {
		msg := errs.Index(i).FieldByName("Msg").String()
		positions := errs.Index(i).FieldByName("Positions")
		m := positions.Len()
		for j := 0; j < m; j++ {
			sourceLine := positions.Index(j).
				FieldByName("TemplateLine").Elem().
				FieldByName("SourceLine")

			if sourceLine.IsNil() {
				sourceLine = positions.Index(j).
					FieldByName("BeforeTemplateLine").Elem().
					FieldByName("SourceLine")
			}
			pos := sourceLine.Elem().FieldByName("Position").Elem()
			errors = append(errors, LinterError{
				Msg: msg,
				Pos: fmt.Sprintf("%s:%d", pos.FieldByName("file").String(), pos.FieldByName("line").Elem().Int()),
			})
		}
	}
	return errors
}