package yttlint

import (
	"github.com/k14s/ytt/pkg/template/core"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type magicType struct {
	CouldBeString bool
	CouldBeInt    bool
	CouldBeFloat  bool
}

func (mt *magicType) Freeze() {
}

func (mt *magicType) Attr(name string) (starlark.Value, error) {
	return &magicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}, nil
}

func (mt *magicType) AttrNames() []string {
	return []string{"data"}
}

func (mt *magicType) String() string {
	return mt.Type()
}
func (mt *magicType) Type() string {
	return "magic"
}
func (mt *magicType) Hash() (uint32, error) {
	return 1, nil
}
func (mt *magicType) Truth() starlark.Bool {
	return starlark.True
}

var _ starlark.HasAttrs = &magicType{}

func (mt *magicType) Name() string {
	return "magic"
}

func (mt *magicType) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	//fmt.Println("CallInternal")
	return &magicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}, nil
}

var _ starlark.Callable = &magicType{}

func (mt *magicType) Iterate() starlark.Iterator {
	return starlark.NewList([]starlark.Value{&magicType{}}).Iterate()
}

var _ starlark.Iterable = &magicType{}

var _ starlark.HasBinary = &magicType{}

func (mt *magicType) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
	_, isString := y.(starlark.String)

	if op == syntax.PLUS && isString { // concatenating any value with a string results in a string
		return &magicType{
			CouldBeString: true,
			CouldBeInt:    false,
			CouldBeFloat:  false,
		}, nil
	}

	return &magicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}, nil
}

func (mt *magicType) AsGoValue() interface{} {
	return mt
}

var _ core.StarlarkValueToGoValueConversion = &magicType{}
