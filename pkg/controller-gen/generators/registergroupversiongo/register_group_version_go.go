package registergroupversiongo

import (
	"fmt"
	"io"
	"strings"

	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"github.com/rancher/wrangler/v2/pkg/name"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
)

type Args struct {
	TypesByGroup map[schema.GroupVersion][]*types.Name
}

func RegisterGroupVersionGo(gv schema.GroupVersion, args *Args) generator.Generator {
	return &registerGroupVersionGo{
		gv:   gv,
		args: args,
		GoGenerator: generator.GoGenerator{
			OutputFilename: "zz_generated_register",
		},
	}
}

type registerGroupVersionGo struct {
	generator.GoGenerator

	gv   schema.GroupVersion
	args *Args
}

func (f *registerGroupVersionGo) Imports(*generator.Context) []string {
	firstType := f.args.TypesByGroup[f.gv][0]
	typeGroupPath := strings.TrimSuffix(firstType.Package, "/"+f.gv.Version)

	packages := append(util.Imports,
		fmt.Sprintf("%s \"%s\"", util.GroupPath(f.gv.Group), typeGroupPath))

	return packages
}

func (f *registerGroupVersionGo) Init(c *generator.Context, w io.Writer) error {
	var (
		types   []*types.Type
		orderer = namer.Orderer{Namer: namer.NewPrivateNamer(0)}
		sw      = generator.NewSnippetWriter(w, c, "{{", "}}")
	)

	for _, name := range f.args.TypesByGroup[f.gv] {
		types = append(types, c.Universe.Type(*name))
	}
	types = orderer.OrderTypes(types)

	m := map[string]interface{}{
		"version":   f.gv.Version,
		"groupPath": util.GroupPath(f.gv.Group),
	}

	sw.Do("var (\n", nil)
	for _, t := range types {
		m := map[string]interface{}{
			"name":   t.Name.Name + "ResourceName",
			"plural": name.GuessPluralName(strings.ToLower(t.Name.Name)),
		}

		sw.Do("{{.name}} = \"{{.plural}}\"\n", m)
	}
	sw.Do(")\n", nil)

	sw.Do(registerGroupVersionBody, m)

	for _, t := range types {
		m := map[string]interface{}{
			"type": t.Name.Name,
		}

		sw.Do("&{{.type}}{},\n", m)
		sw.Do("&{{.type}}List{},\n", m)
	}

	sw.Do(registerGroupVersionBodyEnd, nil)

	return sw.Error()
}

var registerGroupVersionBody = `
// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: {{.groupPath}}.GroupName, Version: "{{.version}}"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
`

var registerGroupVersionBodyEnd = `
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
`

func (f *registerGroupVersionGo) FileType() string { return "" }

func (f *registerGroupVersionGo) Filename() string { return "" }

func (f *registerGroupVersionGo) Filter(ctx *generator.Context, t *types.Type) bool { return false }

func (f *registerGroupVersionGo) Finalize(ctx *generator.Context, ioWriter io.Writer) error {
	return nil
}

func (f *registerGroupVersionGo) GenerateType(ctx *generator.Context, t *types.Type, ioWriter io.Writer) error {
	return nil
}

func (f *registerGroupVersionGo) Name() string {
	return ""
}

func (f *registerGroupVersionGo) Namers(ctx *generator.Context) namer.NameSystems {
	return namer.NameSystems{}
}

func (f *registerGroupVersionGo) PackageVars(ctx *generator.Context) []string {
	return []string{}
}

func (f *registerGroupVersionGo) PackageConsts(*generator.Context) []string {
	return nil
}
