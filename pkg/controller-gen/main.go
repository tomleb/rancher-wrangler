package controllergen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/gengo/v2"
	"k8s.io/gengo/v2/generator"

	cgargs "github.com/rancher/wrangler/v2/pkg/controller-gen/args"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/factorygo"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/groupinterfacego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/groupversioninterfacego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/listtypego"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/registergroupversiongo"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/generators/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	csargs "k8s.io/code-generator/cmd/client-gen/args"
	// clientgenerators "k8s.io/code-generator/cmd/client-gen/generators"
	cs "k8s.io/code-generator/cmd/client-gen/generators"
	types2 "k8s.io/code-generator/cmd/client-gen/types"
	// dpargs "k8s.io/code-generator/cmd/deepcopy-gen/args"
	// dp "k8s.io/code-generator/cmd/deepcopy-gen/generators"
	// infargs "k8s.io/code-generator/cmd/informer-gen/args"
	// inf "k8s.io/code-generator/cmd/informer-gen/generators"
	// lsargs "k8s.io/code-generator/cmd/lister-gen/args"
	// ls "k8s.io/code-generator/cmd/lister-gen/generators"

	"k8s.io/gengo/v2/types"
)

func Run(opts cgargs.Options) {
	args := &generators.Args{
		ImportPackage:    opts.ImportPackage,
		Options:          opts,
		TypesByGroup:     map[schema.GroupVersion][]*types.Name{},
		Package:          opts.OutputPackage,
		GoHeaderFilePath: opts.Boilerplate,
		OutputBase:       util.DefaultSourceTree(),
	}
	args.InputDirs = parseTypes(opts.Groups, args.TypesByGroup)

	args.FactoryGoArgs = &factorygo.Args{
		TypesByGroup: args.TypesByGroup,
	}
	args.GroupInfArgs = &groupinterfacego.Args{
		Options:       args.Options,
		Package:       args.Package,
		ImportPackage: args.ImportPackage,
		TypesByGroup:  args.TypesByGroup,
	}
	args.GroupVerInfArgs = &groupversioninterfacego.Args{
		TypesByGroup: args.TypesByGroup,
	}
	args.ListTypesArgs = &listtypego.Args{
		TypesByGroup: args.TypesByGroup,
	}

	args.RegGroupVerArgs = &registergroupversiongo.Args{
		TypesByGroup: args.TypesByGroup,
	}

	if args.OutputBase == "./" { //go modules
		tempDir, err := os.MkdirTemp("", "")
		if err != nil {
			return
		}

		args.OutputBase = tempDir
		defer os.RemoveAll(tempDir)
	}

	/*new code ^ ^ */
	// customArgs := &cgargs.CustomArgs{
	// 	ImportPackage: opts.ImportPackage,
	// 	Options:       opts,
	// 	TypesByGroup:  map[schema.GroupVersion][]*types.Name{},
	// 	Package:       opts.OutputPackage,
	// }
	// genericArgs := args.Default().WithoutDefaultFlagParsing()
	// genericArgs.CustomArgs = customArgs
	// genericArgs.GoHeaderFilePath = opts.Boilerplate
	// genericArgs.InputDirs = parseTypes(customArgs)

	// if genericArgs.OutputBase == "./" { //go modules
	// 	tempDir, err := os.MkdirTemp("", "")
	// 	if err != nil {
	// 		return
	// 	}

	// 	genericArgs.OutputBase = tempDir
	// 	defer os.RemoveAll(tempDir)
	// }
	// customArgs.OutputBase = genericArgs.OutputBase

	// 	clientGen := generators.NewClientGenerator(args)
	//
	// 	if err := gengo.Execute(
	// 		clientgenerators.NameSystems(nil),
	// 		clientgenerators.DefaultNameSystem(),
	// 		clientGen.GetTargets, "", nil); err != nil {
	// 		logrus.Fatalf("Error: %v", err)
	// 	}

	groups := map[string]bool{}
	listerGroups := map[string]bool{}
	informerGroups := map[string]bool{}
	deepCopygroups := map[string]bool{}
	for groupName, group := range args.Options.Groups {
		if group.GenerateTypes {
			deepCopygroups[groupName] = true
		}
		if group.GenerateClients {
			groups[groupName] = true
		}
		if group.GenerateListers {
			listerGroups[groupName] = true
		}
		if group.GenerateInformers {
			informerGroups[groupName] = true
		}
	}

	if len(deepCopygroups) == 0 && len(groups) == 0 && len(listerGroups) == 0 && len(informerGroups) == 0 {
		if err := copyGoPathToModules(args); err != nil {
			logrus.Fatalf("go modules copy failed: %v", err)
		}
		return
	}

	if err := copyGoPathToModules(args); err != nil {
		logrus.Fatalf("go modules copy failed: %v", err)
	}

	// if err := generateDeepcopy(deepCopygroups, args); err != nil {
	// 	logrus.Fatalf("deepcopy failed: %v", err)
	// }

	if err := generateClientset(groups, args); err != nil {
		logrus.Fatalf("clientset failed: %v", err)
	}

	// if err := generateListers(listerGroups, args); err != nil {
	// 	logrus.Fatalf("listers failed: %v", err)
	// }
	//
	// if err := generateInformers(informerGroups, args); err != nil {
	// 	logrus.Fatalf("informers failed: %v", err)
	// }
	//
	if err := copyGoPathToModules(args); err != nil {
		logrus.Fatalf("go modules copy failed: %v", err)
	}
}

func sourcePackagePath(customArgs *generators.Args, pkgName string) string {
	pkgSplit := strings.Split(pkgName, string(os.PathSeparator))
	pkg := filepath.Join(customArgs.OutputBase, strings.Join(pkgSplit[:3], string(os.PathSeparator)))
	return pkg
}

// until k8s code-gen supports gopath
func copyGoPathToModules(customArgs *generators.Args) error {

	pathsToCopy := map[string]bool{}
	for _, types := range customArgs.TypesByGroup {
		for _, names := range types {
			pkg := sourcePackagePath(customArgs, names.Package)
			pathsToCopy[pkg] = true
		}
	}

	pkg := sourcePackagePath(customArgs, customArgs.Package)
	pathsToCopy[pkg] = true

	for pkg := range pathsToCopy {
		if _, err := os.Stat(pkg); os.IsNotExist(err) {
			continue
		}

		return filepath.Walk(pkg, func(path string, info os.FileInfo, err error) error {
			newPath := strings.Replace(path, pkg, ".", 1)
			if info.IsDir() {
				return os.MkdirAll(newPath, info.Mode())
			}

			return copyFile(path, newPath)
		})
	}

	return nil
}

func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

/*
func generateDeepcopy(groups map[string]bool, args *generators.Args) error {
	if len(groups) == 0 {
		return nil
	}

	deepCopyCustomArgs := &dpargs.Args{}
	args.OutputBase = "zz_generated_deepcopy"

	for gv, names := range args.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
		deepCopyCustomArgs.BoundingDirs = append(deepCopyCustomArgs.BoundingDirs, names[0].Package)
	}
	targets := setGenClient(groups, args.TypesByGroup, dp.GetTargets)
	return gengo.Execute(dp.NameSystems(),
		dp.DefaultNameSystem(),
		func(ctx *generator.Context) []generator.Target {
			return setGenClient(groups, args.TypesByGroup, dp.GetTargets)(ctx, args)
		}, gengo.StdBuildTag, nil)
}
*/

func generateClientset(groups map[string]bool, args *generators.Args) error {
	if len(groups) == 0 {
		return nil
	}

	clientSetArgs := csargs.New()
	clientSetArgs.ClientsetName = "versioned"
	clientSetArgs.OutputDir = args.OutputBase
	clientSetArgs.OutputPkg = filepath.Join(args.Package)
	clientSetArgs.GoHeaderFile = args.Options.Boilerplate
	// args.OutputPackagePath = filepath.Join(args.Package, "clientset")

	var order []schema.GroupVersion

	for gv := range args.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		order = append(order, gv)
	}
	sort.Slice(order, func(i, j int) bool {
		return order[i].Group < order[j].Group
	})

	inputDirs := []string{}
	for _, gv := range order {
		packageName := args.Options.Groups[gv.Group].PackageName
		if packageName == "" {
			packageName = gv.Group
		}
		names := args.TypesByGroup[gv]
		inputDirs = append(inputDirs, names[0].Package)
		clientSetArgs.Groups = append(clientSetArgs.Groups, types2.GroupVersions{
			PackageName: packageName,
			Group:       types2.Group(gv.Group),
			Versions: []types2.PackageVersion{
				{
					Version: types2.Version(gv.Version),
					Package: names[0].Package,
				},
			},
		})
	}
	getTargets := setGenClient(groups, args.TypesByGroup, func(context *generator.Context) []generator.Target {
		return cs.GetTargets(context, clientSetArgs)
	})
	return gengo.Execute(cs.NameSystems(nil), cs.DefaultNameSystem(), getTargets, gengo.StdBuildTag, inputDirs)
}

/*
func generateInformers(groups map[string]bool, args *generators.Args) error {
	if len(groups) == 0 {
		return nil
	}

	args, clientSetArgs := infargs.NewDefaults()
	clientSetArgs.VersionedClientSetPackage = filepath.Join(customArgs.Package, "clientset/versioned")
	clientSetArgs.ListersPackage = filepath.Join(customArgs.Package, "listers")
	args.OutputBase = customArgs.OutputBase
	args.OutputPackagePath = filepath.Join(customArgs.Package, "informers")
	args.GoHeaderFilePath = customArgs.Options.Boilerplate

	for gv, names := range customArgs.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
	}

	getTargets := setGenClient(groups, args.TypesByGroup, func(context *generator.Context) []generator.Target {
		return cs.GetTargets(context, clientSetArgs)
	})

	return gengo.Execute(inf.NameSystems(nil),
		inf.DefaultNameSystem(),
		setGenClient(groups, customArgs.TypesByGroup, inf.Packages))
}

func generateListers(groups map[string]bool, args *generators.Args) error {
	if len(groups) == 0 {
		return nil
	}

	args, _ := lsargs.NewDefaults()
	args.OutputBase = customArgs.OutputBase
	args.OutputPackagePath = filepath.Join(customArgs.Package, "listers")
	args.GoHeaderFilePath = customArgs.Options.Boilerplate

	for gv, names := range customArgs.TypesByGroup {
		if !groups[gv.Group] {
			continue
		}
		args.InputDirs = append(args.InputDirs, names[0].Package)
	}

	return args.Execute(ls.NameSystems(nil),
		ls.DefaultNameSystem(),
		setGenClient(groups, customArgs.TypesByGroup, ls.Packages))
}
*/

func setGenClient(
	groups map[string]bool,
	typesByGroup map[schema.GroupVersion][]*types.Name,
	f func(*generator.Context) []generator.Target,
) func(*generator.Context) []generator.Target {
	return func(context *generator.Context) []generator.Target {
		for gv, names := range typesByGroup {
			if !groups[gv.Group] {
				continue
			}
			for _, name := range names {
				var (
					p           = context.Universe.Package(name.Package)
					t           = p.Type(name.Name)
					status      bool
					nsed        bool
					kubebuilder bool
				)

				for _, line := range append(t.SecondClosestCommentLines, t.CommentLines...) {
					switch {
					case strings.Contains(line, "+kubebuilder:object:root=true"):
						kubebuilder = true
						t.SecondClosestCommentLines = append(t.SecondClosestCommentLines, "+genclient")
					case strings.Contains(line, "+kubebuilder:subresource:status"):
						status = true
					case strings.Contains(line, "+kubebuilder:resource:") && strings.Contains(line, "scope=Namespaced"):
						nsed = true
					}
				}

				if kubebuilder {
					if !nsed {
						t.SecondClosestCommentLines = append(t.SecondClosestCommentLines, "+genclient:nonNamespaced")
					}
					if !status {
						t.SecondClosestCommentLines = append(t.SecondClosestCommentLines, "+genclient:noStatus")
					}

					foundGroup := false
					for _, comment := range p.DocComments {
						if strings.Contains(comment, "+groupName=") {
							foundGroup = true
							break
						}
					}

					if !foundGroup {
						p.DocComments = append(p.DocComments, "+groupName="+gv.Group)
						p.Comments = append(p.Comments, "+groupName="+gv.Group)
						fmt.Println(gv.Group, p.DocComments, p.Comments, p.Path)
					}
				}
			}
		}
		return f(context)
	}
}

func parseTypes(groups map[string]cgargs.Group, typesByGroup map[schema.GroupVersion][]*types.Name) []string {
	for groupName, group := range groups {
		if group.GenerateTypes || group.GenerateClients {
			groups[groupName] = group
		}
	}

	for groupName, group := range groups {
		if err := cgargs.ObjectsToGroupVersion(groupName, group.Types, typesByGroup); err != nil {
			// sorry, should really handle this better
			panic(err)
		}
	}

	var inputDirs []string
	for _, names := range typesByGroup {
		inputDirs = append(inputDirs, names[0].Package)
	}

	return inputDirs
}
