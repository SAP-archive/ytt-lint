package yttlint

import "fmt"

type LinterError struct {
	Msg string `json:"msg"`
	Pos string `json:"pos"`
}

func lintErrorf(format string, args ...interface{}) LinterError {
	return LinterError{
		Msg: fmt.Sprintf(format, args...),
	}
}
