package librarywrapper

import (
	"fmt"

	"github.com/k14s/ytt/pkg/template/core"
	"github.com/k14s/ytt/pkg/yttlibrary"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/SAP/ytt-lint/pkg/magic"
)

var (
	Base64APIWrapper = starlark.StringDict{
		"base64": &starlarkstruct.Module{
			Name: "base64",
			Members: starlark.StringDict{
				"encode": starlark.NewBuiltin("base64.encode", core.ErrWrapper(base64Module{}.Encode)),
				"decode": starlark.NewBuiltin("base64.decode", core.ErrWrapper(base64Module{}.Decode)),
			},
		},
	}
	base64module = yttlibrary.Base64API["base64"].(*starlarkstruct.Module)
)

func init() {
	yttlibrary.Base64API = Base64APIWrapper
}

type base64Module struct{}

func (b base64Module) Encode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	_, ok := args.Index(0).(*magic.MagicType)
	if ok {
		return &magic.MagicType{
			CouldBeString: true,
			CouldBeInt:    false,
			CouldBeFloat:  false,
		}, nil
	}

	encode, err := base64module.Attr("encode")
	if err != nil {
		return starlark.None, err
	}
	return encode.(*starlark.Builtin).CallInternal(thread, args, kwargs)
}

func (b base64Module) Decode(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	_, ok := args.Index(0).(*magic.MagicType)
	if ok {
		return &magic.MagicType{
			CouldBeString: true,
			CouldBeInt:    false,
			CouldBeFloat:  false,
		}, nil
	}

	decode, err := base64module.Attr("decode")
	if err != nil {
		return starlark.None, err
	}
	return decode.(*starlark.Builtin).CallInternal(thread, args, kwargs)
}
