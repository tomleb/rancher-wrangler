package factorygo

import (
	"fmt"
	"io"

	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
)

type Args struct {
	TypesByGroup map[schema.GroupVersion][]*types.Name
}

func FactoryGo(group string, args *Args) generator.Generator {
	return &factory{
		group: group,
		args:  args,
		GoGenerator: generator.GoGenerator{
			OutputFilename: "factory",
			OptionalBody:   []byte(factoryBody),
		},
	}
}

type factory struct {
	generator.GoGenerator

	group string
	args  *Args
}

func (f *factory) Imports(*generator.Context) []string {
	imports := util.Imports

	for gv, types := range f.args.TypesByGroup {
		if f.group == gv.Group && len(types) > 0 {
			imports = append(imports,
				fmt.Sprintf("%s \"%s\"", gv.Version, types[0].Package))
		}
	}

	return imports
}

func (f *factory) Init(c *generator.Context, w io.Writer) error {
	if err := f.GoGenerator.Init(c, w); err != nil {
		return err
	}

	sw := generator.NewSnippetWriter(w, c, "{{", "}}")
	m := map[string]interface{}{
		"groupName": util.UpperLowercase(f.group),
	}

	sw.Do("\n\nfunc (c *Factory) {{.groupName}}() Interface {\n", m)
	sw.Do("	return New(c.ControllerFactory())\n", m)
	sw.Do("}\n\n", m)

	sw.Do("\n\nfunc (c *Factory) WithAgent(userAgent string) Interface {\n", m)
	sw.Do("	return New(controller.NewSharedControllerFactoryWithAgent(userAgent, c.ControllerFactory()))\n", m)
	sw.Do("}\n\n", m)

	return sw.Error()
}

var factoryBody = `
type Factory struct {
	*generic.Factory
}

func NewFactoryFromConfigOrDie(config *rest.Config) *Factory {
	f, err := NewFactoryFromConfig(config)
	if err != nil {
		panic(err)
	}
	return f
}

func NewFactoryFromConfig(config *rest.Config) (*Factory, error) {
	return NewFactoryFromConfigWithOptions(config, nil)
}

func NewFactoryFromConfigWithNamespace(config *rest.Config, namespace string) (*Factory, error) {
	return NewFactoryFromConfigWithOptions(config, &FactoryOptions{
		Namespace: namespace,
	})
}

type FactoryOptions = generic.FactoryOptions

func NewFactoryFromConfigWithOptions(config *rest.Config, opts *FactoryOptions) (*Factory, error) {
	f, err := generic.NewFactoryFromConfigWithOptions(config, opts)
	return &Factory{
		Factory: f,
	}, err
}

func NewFactoryFromConfigWithOptionsOrDie(config *rest.Config, opts *FactoryOptions) *Factory {
    f, err := NewFactoryFromConfigWithOptions(config, opts)
	if err != nil {
		panic(err)
	}
	return f
}

`

func (f *factory) FileType() string { return "" }

func (f *factory) Filename() string { return "" }

func (f *factory) Filter(ctx *generator.Context, t *types.Type) bool { return false }

func (f *factory) Finalize(ctx *generator.Context, ioWriter io.Writer) error { return nil }

func (f *factory) GenerateType(ctx *generator.Context, t *types.Type, ioWriter io.Writer) error {
	return nil
}

func (f *factory) Name() string {
	return ""
}

func (f *factory) Namers(ctx *generator.Context) namer.NameSystems {
	return namer.NameSystems{}
}

func (f *factory) PackageVars(ctx *generator.Context) []string {
	return []string{}
}

func (f *factory) PackageConsts(*generator.Context) []string {
	return []string{
		fmt.Sprintf("GroupName = \"%s\"", f.group),
	}
}
