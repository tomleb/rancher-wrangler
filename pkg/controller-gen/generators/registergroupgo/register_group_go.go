package registergroupgo

import (
	"fmt"
	"io"

	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
)

type Args struct {
}

func RegisterGroupGo(group string, args *Args) generator.Generator {
	return &registerGroupGo{
		group: group,
		args:  args,
		GoGenerator: generator.GoGenerator{
			OutputFilename: "zz_generated_register",
		},
	}
}

type registerGroupGo struct {
	generator.GoGenerator

	group string
	args  *Args
}

func (f *registerGroupGo) PackageConsts(*generator.Context) []string {
	return []string{
		fmt.Sprintf("GroupName = \"%s\"", f.group),
	}
}

func (f *registerGroupGo) FileType() string { return "" }

func (f *registerGroupGo) Filename() string { return "" }

func (f *registerGroupGo) Filter(ctx *generator.Context, t *types.Type) bool { return false }

func (f *registerGroupGo) Finalize(ctx *generator.Context, ioWriter io.Writer) error { return nil }

func (f *registerGroupGo) GenerateType(ctx *generator.Context, t *types.Type, ioWriter io.Writer) error {
	return nil
}

func (f *registerGroupGo) Imports(ctx *generator.Context) []string {
	return nil
}

func (f *registerGroupGo) Init(ctx *generator.Context, ioWriter io.Writer) error {
	return nil
}

func (f *registerGroupGo) Name() string {
	return ""
}

func (f *registerGroupGo) Namers(ctx *generator.Context) namer.NameSystems {
	return namer.NameSystems{}
}

func (f *registerGroupGo) PackageVars(ctx *generator.Context) []string {
	return []string{}
}
