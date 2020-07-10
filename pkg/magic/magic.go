package magic

import (
	"github.com/k14s/ytt/pkg/template/core"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type MagicType struct {
	CouldBeString bool
	CouldBeInt    bool
	CouldBeFloat  bool
}

func (mt *MagicType) Freeze() {
}

func (mt *MagicType) Attr(name string) (starlark.Value, error) {
	return &MagicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}, nil
}

func (mt *MagicType) AttrNames() []string {
	return []string{"data"}
}

func (mt *MagicType) String() string {
	return mt.Type()
}
func (mt *MagicType) Type() string {
	return "magic"
}
func (mt *MagicType) Hash() (uint32, error) {
	return 1, nil
}
func (mt *MagicType) Truth() starlark.Bool {
	return starlark.True
}

var _ starlark.HasAttrs = &MagicType{}

func (mt *MagicType) Name() string {
	return "magic"
}

func (mt *MagicType) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	//fmt.Println("CallInternal")
	return &MagicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}, nil
}

var _ starlark.Callable = &MagicType{}

func (mt *MagicType) Iterate() starlark.Iterator {
	return starlark.NewList([]starlark.Value{&MagicType{}}).Iterate()
}

var _ starlark.Iterable = &MagicType{}

func (mt *MagicType) Len() int {
	return 42
}

var _ starlark.Sequence = &MagicType{}

var _ starlark.HasBinary = &MagicType{}

func (mt *MagicType) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
	_, isString := y.(starlark.String)

	if op == syntax.PLUS && isString { // concatenating any value with a string results in a string
		return &MagicType{
			CouldBeString: true,
			CouldBeInt:    false,
			CouldBeFloat:  false,
		}, nil
	}

	return &MagicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}, nil
}

var _ starlark.Sliceable = &MagicType{}

func (mt *MagicType) Slice(start, end, step int) starlark.Value {
	return &MagicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}
}

func (mt *MagicType) Index(i int) starlark.Value {
	return &MagicType{
		CouldBeString: true,
		CouldBeInt:    true,
		CouldBeFloat:  true,
	}
}

func (mt *MagicType) AsGoValue() interface{} {
	return mt
}

func (mt *MagicType) AsStarlarkValue() starlark.Value {
	return mt
}

var _ core.StarlarkValueToGoValueConversion = &MagicType{}
var _ core.GoValueToStarlarkValueConversion = &MagicType{}
