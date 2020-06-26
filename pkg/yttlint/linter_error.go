package yttlint

import "fmt"

type ErrorCode string

const (
	ErrorCodeHelm = "HELM"
)

type LinterError struct {
	Msg  string    `json:"msg"`
	Pos  string    `json:"pos"`
	Code ErrorCode `json:"code"`
}

func lintErrorf(format string, args ...interface{}) LinterError {
	return LinterError{
		Msg: fmt.Sprintf(format, args...),
	}
}
