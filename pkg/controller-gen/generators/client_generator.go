/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generators

import (
	"fmt"
	"path/filepath"
	"strings"

	args2 "github.com/rancher/wrangler/v2/pkg/controller-gen/args"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/factorygo"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/groupinterfacego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/groupversioninterfacego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/listtypego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/registergroupgo"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/registergroupversiongo"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/typego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/types"
)

type Args struct {
	GoHeaderFilePath  string
	Options           args2.Options
	TypesByGroup      map[schema.GroupVersion][]*types.Name
	Package           string
	ImportPackage     string
	InputDirs         []string
	OutputBase        string
	OutputPackagePath string

	FactoryGoArgs   *factorygo.Args
	GroupInfArgs    *groupinterfacego.Args
	GroupVerInfArgs *groupversioninterfacego.Args
	ListTypesArgs   *listtypego.Args
	RegGroupArgs    *registergroupgo.Args
	RegGroupVerArgs *registergroupversiongo.Args
	TypeGoArgs      *typego.Args
}

type ClientGenerator struct {
	Fakes map[string][]string
	Args  *Args
}

func NewClientGenerator(args *Args) *ClientGenerator {
	return &ClientGenerator{
		Fakes: make(map[string][]string),
		Args:  args,
	}
}

// Packages makes the client package definition.
func (cg *ClientGenerator) GetTargets(context *generator.Context) []generator.Target {
	generateTypesGroups := map[string]bool{}

	for groupName, group := range cg.Args.Options.Groups {
		if group.GenerateTypes {
			generateTypesGroups[groupName] = true
		}
	}

	var (
		packageList []generator.Target
		groups      = map[string]bool{}
	)

	for gv, types := range cg.Args.TypesByGroup {
		if !groups[gv.Group] {
			packageList = append(packageList, cg.groupPackage(gv.Group))
			if generateTypesGroups[gv.Group] {
				packageList = append(packageList, cg.typesGroupPackage(types[0], gv))
			}
		}
		groups[gv.Group] = true
		packageList = append(packageList, cg.groupVersionPackage(gv))

		if generateTypesGroups[gv.Group] {
			packageList = append(packageList, cg.typesGroupVersionPackage(types[0], gv))
			packageList = append(packageList, cg.typesGroupVersionDocPackage(types[0], gv))
		}
	}

	return packageList
}

func (cg *ClientGenerator) typesGroupPackage(name *types.Name, gv schema.GroupVersion) generator.Target {
	packagePath := strings.TrimSuffix(name.Package, "/"+gv.Version)
	return Target(cg.Args.GoHeaderFilePath, packagePath, func(context *generator.Context) []generator.Generator {
		return []generator.Generator{
			registergroupgo.RegisterGroupGo(gv.Group, cg.Args.RegGroupArgs),
		}
	})
}

func (cg *ClientGenerator) typesGroupVersionDocPackage(name *types.Name, gv schema.GroupVersion) generator.Target {
	packagePath := name.Package
	p := Target(cg.Args.GoHeaderFilePath, packagePath, func(context *generator.Context) []generator.Generator {
		return []generator.Generator{
			generator.GoGenerator{
				OutputFilename: "doc",
			},
			registergroupversiongo.RegisterGroupVersionGo(gv, cg.Args.RegGroupVerArgs),
			listtypego.ListTypesGo(gv, cg.Args.ListTypesArgs),
		}
	})

	p.(*generator.SimpleTarget).HeaderComment = append(p.(*generator.SimpleTarget).HeaderComment, []byte(fmt.Sprintf(`

// +k8s:deepcopy-gen=package
// +groupName=%s
`, gv.Group))...)

	return p
}

func (cg *ClientGenerator) typesGroupVersionPackage(name *types.Name, gv schema.GroupVersion) generator.Target {
	packagePath := name.Package
	return Target(cg.Args.GoHeaderFilePath, packagePath, func(context *generator.Context) []generator.Generator {
		return []generator.Generator{
			registergroupversiongo.RegisterGroupVersionGo(gv, cg.Args.RegGroupVerArgs),
			listtypego.ListTypesGo(gv, cg.Args.ListTypesArgs),
		}
	})
}

func (cg *ClientGenerator) groupPackage(group string) generator.Target {
	packagePath := filepath.Join(cg.Args.Package, "controllers", util.GroupPackageName(group, cg.Args.Options.Groups[group].OutputControllerPackageName))

	return Target(cg.Args.GoHeaderFilePath, packagePath, func(context *generator.Context) []generator.Generator {
		return []generator.Generator{
			factorygo.FactoryGo(group, cg.Args.FactoryGoArgs),
			groupinterfacego.GroupInterfaceGo(group, cg.Args.GroupInfArgs),
		}
	})
}

func (cg *ClientGenerator) groupVersionPackage(gv schema.GroupVersion) generator.Target {
	packagePath := filepath.Join(cg.Args.Package, "controllers", util.GroupPackageName(gv.Group, cg.Args.Options.Groups[gv.Group].OutputControllerPackageName), gv.Version)

	return Target(cg.Args.GoHeaderFilePath, packagePath, func(context *generator.Context) []generator.Generator {
		generators := []generator.Generator{
			groupversioninterfacego.GroupVersionInterfaceGo(gv, cg.Args.GroupVerInfArgs),
		}

		for _, t := range cg.Args.TypesByGroup[gv] {
			generators = append(generators, typego.TypeGo(gv, t, cg.Args.TypeGoArgs))
			cg.Fakes[packagePath] = append(cg.Fakes[packagePath], t.Name)
		}

		return generators
	})
}
