package generators

import (
	"strings"

	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/gengo/v2"
	"k8s.io/gengo/v2/generator"
)

func Target(headerFile, name string, generators func(context *generator.Context) []generator.Generator) generator.Target {
	boilerplate, err := gengo.GoBoilerplate(headerFile, gengo.StdBuildTag, gengo.StdGeneratedBy)
	runtime.Must(err)

	parts := strings.Split(name, "/")
	return &generator.SimpleTarget{
		PkgName:        util.GroupPath(parts[len(parts)-1]),
		PkgPath:        name,
		HeaderComment:  boilerplate,
		GeneratorsFunc: generators,
	}
}
