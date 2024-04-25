package groupversioninterfacego

import (
	"fmt"
	"io"
	"strings"

	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
)

type Args struct {
	TypesByGroup map[schema.GroupVersion][]*types.Name
}

func GroupVersionInterfaceGo(gv schema.GroupVersion, args *Args) generator.Generator {
	return &groupInterfaceGo{
		gv:   gv,
		args: args,
		GoGenerator: generator.GoGenerator{
			OutputFilename: "interface",
		},
	}
}

type groupInterfaceGo struct {
	generator.GoGenerator

	gv   schema.GroupVersion
	args *Args
}

func (f *groupInterfaceGo) Imports(context *generator.Context) []string {
	firstType := f.args.TypesByGroup[f.gv][0]

	packages := append(util.Imports,
		fmt.Sprintf("%s \"%s\"", f.gv.Version, firstType.Package))

	return packages
}

var (
	pluralExceptions = map[string]string{
		"Endpoints": "Endpoints",
	}
	plural = namer.NewPublicPluralNamer(pluralExceptions)
)

func (f *groupInterfaceGo) Init(c *generator.Context, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")

	orderer := namer.Orderer{Namer: namer.NewPrivateNamer(0)}

	var types []*types.Type
	for _, name := range f.args.TypesByGroup[f.gv] {
		types = append(types, c.Universe.Type(*name))
	}
	types = orderer.OrderTypes(types)

	sw.Do("func init() {\n", nil)
	sw.Do("schemes.Register("+f.gv.Version+".AddToScheme)\n", nil)
	sw.Do("}\n", nil)

	sw.Do("type Interface interface {\n", nil)
	for _, t := range types {
		m := map[string]interface{}{
			"type": t.Name.Name,
		}
		sw.Do("{{.type}}() {{.type}}Controller\n", m)
	}
	sw.Do("}\n", nil)

	m := map[string]interface{}{
		"version":      f.gv.Version,
		"versionUpper": namer.IC(f.gv.Version),
		"groupUpper":   util.UpperLowercase(f.gv.Group),
	}
	sw.Do(groupInterfaceBody, m)

	for _, t := range types {
		m := map[string]interface{}{
			"type":         t.Name.Name,
			"plural":       plural.Name(t),
			"pluralLower":  strings.ToLower(plural.Name(t)),
			"version":      f.gv.Version,
			"group":        f.gv.Group,
			"namespaced":   util.Namespaced(t),
			"versionUpper": namer.IC(f.gv.Version),
			"groupUpper":   util.UpperLowercase(f.gv.Group),
		}
		body := `
		func (v *version) {{.type}}() {{.type}}Controller {
			return generic.New{{ if not .namespaced}}NonNamespaced{{end}}Controller[*{{.version}}.{{.type}}, *{{.version}}.{{.type}}List](schema.GroupVersionKind{Group: "{{.group}}", Version: "{{.version}}", Kind: "{{.type}}"}, "{{.pluralLower}}", {{ if .namespaced}}true, {{end}}v.controllerFactory)
		}
		`
		sw.Do(body, m)
	}

	return sw.Error()
}

var groupInterfaceBody = `
func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &version{
		controllerFactory: controllerFactory,
	}
}

type version struct {
	controllerFactory controller.SharedControllerFactory
}

`

func (f *groupInterfaceGo) FileType() string { return "" }

func (f *groupInterfaceGo) Filename() string { return "" }

func (f *groupInterfaceGo) Filter(ctx *generator.Context, t *types.Type) bool { return false }

func (f *groupInterfaceGo) Finalize(ctx *generator.Context, ioWriter io.Writer) error { return nil }

func (f *groupInterfaceGo) GenerateType(ctx *generator.Context, t *types.Type, ioWriter io.Writer) error {
	return nil
}

func (f *groupInterfaceGo) Name() string {
	return ""
}

func (f *groupInterfaceGo) Namers(ctx *generator.Context) namer.NameSystems {
	return namer.NameSystems{}
}

func (f *groupInterfaceGo) PackageVars(ctx *generator.Context) []string {
	return []string{}
}

func (f *groupInterfaceGo) PackageConsts(*generator.Context) []string {
	return nil
}
