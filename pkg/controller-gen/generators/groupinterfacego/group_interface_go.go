package groupinterfacego

import (
	"fmt"
	"io"

	args2 "github.com/rancher/wrangler/v2/pkg/controller-gen/args"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
)

type Args struct {
	Options       args2.Options
	TypesByGroup  map[schema.GroupVersion][]*types.Name
	Package       string
	ImportPackage string
}

func GroupInterfaceGo(group string, args *Args) generator.Generator {
	return &interfaceGo{
		group: group,
		args:  args,
		GoGenerator: generator.GoGenerator{
			OutputFilename: "interface",
			OptionalBody:   []byte(interfaceBody),
		},
	}
}

type interfaceGo struct {
	generator.GoGenerator

	group string
	args  *Args
}

func (f *interfaceGo) Imports(context *generator.Context) []string {
	packages := util.Imports

	for gv := range f.args.TypesByGroup {
		if gv.Group != f.group {
			continue
		}
		pkg := f.args.ImportPackage
		if pkg == "" {
			pkg = f.args.Package
		}
		packages = append(packages, fmt.Sprintf("%s \"%s/controllers/%s/%s\"", gv.Version, pkg,
			util.GroupPackageName(gv.Group, f.args.Options.Groups[gv.Group].OutputControllerPackageName), gv.Version))
	}

	return packages
}

func (f *interfaceGo) Init(c *generator.Context, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")
	sw.Do("type Interface interface {\n", nil)
	for gv := range f.args.TypesByGroup {
		if gv.Group != f.group {
			continue
		}

		sw.Do("{{.upperVersion}}() {{.version}}.Interface\n", map[string]interface{}{
			"upperVersion": namer.IC(gv.Version),
			"version":      gv.Version,
		})
	}
	sw.Do("}\n", nil)

	if err := f.GoGenerator.Init(c, w); err != nil {
		return err
	}

	for gv := range f.args.TypesByGroup {
		if gv.Group != f.group {
			continue
		}

		m := map[string]interface{}{
			"upperGroup":   util.UpperLowercase(f.group),
			"upperVersion": namer.IC(gv.Version),
			"version":      gv.Version,
		}
		sw.Do("\nfunc (g *group) {{.upperVersion}}() {{.version}}.Interface {\n", m)
		sw.Do("return {{.version}}.New(g.controllerFactory)\n", m)
		sw.Do("}\n", m)
	}

	return sw.Error()
}

var interfaceBody = `
type group struct {
	controllerFactory controller.SharedControllerFactory
}

// New returns a new Interface.
func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &group{
		controllerFactory: controllerFactory,
	}
}
`

func (f *interfaceGo) FileType() string { return "" }

func (f *interfaceGo) Filename() string { return "" }

func (f *interfaceGo) Filter(ctx *generator.Context, t *types.Type) bool { return false }

func (f *interfaceGo) Finalize(ctx *generator.Context, ioWriter io.Writer) error { return nil }

func (f *interfaceGo) GenerateType(ctx *generator.Context, t *types.Type, ioWriter io.Writer) error {
	return nil
}

func (f *interfaceGo) Name() string {
	return ""
}

func (f *interfaceGo) Namers(ctx *generator.Context) namer.NameSystems {
	return namer.NameSystems{}
}

func (f *interfaceGo) PackageVars(ctx *generator.Context) []string {
	return []string{}
}

func (f *interfaceGo) PackageConsts(*generator.Context) []string {
	return []string{
		fmt.Sprintf("GroupName = \"%s\"", f.group),
	}
}
