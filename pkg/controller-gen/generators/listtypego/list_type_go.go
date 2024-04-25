package listtypego

import (
	"io"

	args2 "github.com/rancher/wrangler/v2/pkg/controller-gen/args"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
)

type Args struct {
	TypesByGroup map[schema.GroupVersion][]*types.Name
}

func ListTypesGo(gv schema.GroupVersion, args *Args) generator.Generator {
	return &listTypesGo{
		gv:   gv,
		args: args,
		GoGenerator: generator.GoGenerator{
			OutputFilename: "zz_generated_list_types",
		},
	}
}

type listTypesGo struct {
	generator.GoGenerator

	gv   schema.GroupVersion
	args *Args
}

func (f *listTypesGo) Imports(*generator.Context) []string {
	return util.Imports
}

func (f *listTypesGo) Init(c *generator.Context, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")

	for _, t := range f.args.TypesByGroup[f.gv] {
		m := map[string]interface{}{
			"type": t.Name,
		}
		args2.CheckType(c.Universe.Type(*t))
		sw.Do(string(listTypesBody), m)
	}

	return sw.Error()
}

var listTypesBody = `
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// {{.type}}List is a list of {{.type}} resources
type {{.type}}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata"` + "`" + `

	Items []{{.type}} ` + "`" + `json:"items"` + "`" + `
}

func New{{.type}}(namespace, name string, obj {{.type}}) *{{.type}} {
	obj.APIVersion, obj.Kind = SchemeGroupVersion.WithKind("{{.type}}").ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}
`

func (f *listTypesGo) FileType() string { return "" }

func (f *listTypesGo) Filename() string { return "" }

func (f *listTypesGo) Filter(ctx *generator.Context, t *types.Type) bool { return false }

func (f *listTypesGo) Finalize(ctx *generator.Context, ioWriter io.Writer) error { return nil }

func (f *listTypesGo) GenerateType(ctx *generator.Context, t *types.Type, ioWriter io.Writer) error {
	return nil
}

func (f *listTypesGo) Name() string {
	return ""
}

func (f *listTypesGo) Namers(ctx *generator.Context) namer.NameSystems {
	return namer.NameSystems{}
}

func (f *listTypesGo) PackageVars(ctx *generator.Context) []string {
	return []string{}
}

func (f *listTypesGo) PackageConsts(*generator.Context) []string {
	return nil
}
