package main

import (
	"bytes"
	"cmp"
	"embed"
	_ "embed"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/jessevdk/go-flags"

	"github.com/arikkfir/devbot/internal/util/k8s"
)

var (
	//go:embed *.tmpl
	templatesFS                 embed.FS
	commonConditionsRegexp      = regexp.MustCompile(`\s*\+condition:commons$`)
	conditionRegexp             = regexp.MustCompile(`\s*\+condition:([^:]+),([^:]+):(.*)`)
	conditionTypeRegexp         = regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`)
	reasonRegexp                = regexp.MustCompile(`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`)
	kubeBuilderObjectRootRegexp = regexp.MustCompile(`\s*\+kubebuilder:object:root=true`)
)

type Config struct {
	Args struct {
		Paths []string `positional-arg-name:"paths" description:"Paths to Go files or directories of Go files"`
	} `positional-args:"yes" required:"yes"`
}

type Condition struct {
	Name        string
	RemovalVerb string
	Reasons     []string
}

type Constants struct {
	PackageName string
	Conditions  []Condition
}

func locateFiles(pathGlobs []string) ([]string, error) {
	var files []string
	for _, glob := range pathGlobs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s': %w", glob, err)
		}

		for _, match := range matches {
			err := filepath.WalkDir(match, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(path, ".go") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to traverse '%s': %w", match, err)
			}
		}
	}
	return files, nil
}

func parseObjectConditionTypes(object *ast.Object, commentLines []string) ([]Condition, error) {
	var conditionTypes []Condition
	for _, line := range commentLines {
		if commonConditionsRegexp.MatchString(line) {
			for conditionType, info := range k8s.CommonConditionReasons {
				slices.Sort(info.Reasons)
				conditionTypes = append(conditionTypes,
					Condition{
						Name:        conditionType,
						RemovalVerb: info.RemovalVerb,
						Reasons:     info.Reasons,
					},
				)
			}
		}
	}
	for _, line := range commentLines {
		if matches := conditionRegexp.FindStringSubmatch(line); len(matches) > 0 {
			removalVerb := matches[1]
			if !conditionTypeRegexp.MatchString(removalVerb) {
				return nil, fmt.Errorf("invalid condition verb '%s' for '%s'", removalVerb, object.Name)
			}
			name := matches[2]
			if !conditionTypeRegexp.MatchString(name) {
				return nil, fmt.Errorf("invalid condition type '%s' for '%s'", name, object.Name)
			}
			reasons := strings.Split(matches[3], ",")
			for _, reason := range reasons {
				if !reasonRegexp.MatchString(reason) {
					return nil, fmt.Errorf("invalid reason '%s' for condition '%s' of '%s'", reason, name, object.Name)
				}
			}
			slices.Sort(reasons)
			found := false
			for i, conditionType := range conditionTypes {
				if conditionType.Name == name {
					if conditionType.RemovalVerb != removalVerb {
						return nil, fmt.Errorf("conflicting removal verbs for condition '%s' of '%s'", name, object.Name)
					} else {
						for _, reason := range reasons {
							if !slices.Contains(conditionTypes[i].Reasons, reason) {
								conditionTypes[i].Reasons = append(conditionTypes[i].Reasons, reason)
							}
						}
						slices.Sort(conditionTypes[i].Reasons)
						found = true
						break
					}
				}
			}
			if !found {
				conditionTypes = append(conditionTypes, Condition{
					Name:        name,
					RemovalVerb: removalVerb,
					Reasons:     reasons,
				})
			}
		}
	}
	return conditionTypes, nil
}

func generateConditionsFile(tmpl *template.Template, src string, packageName string, object *ast.Object, conditions []Condition) error {
	genFilename := fmt.Sprintf("%s/zz_generated.%s.conditions.go", filepath.Dir(src), strings.ToLower(object.Name))

	// Sort data to prevent diffs due to unpredictable symbol parsing order
	for _, c := range conditions {
		sort.Slice(c.Reasons, func(i, j int) bool { return cmp.Less(c.Reasons[i], c.Reasons[j]) })
	}
	sort.Slice(conditions, func(i, j int) bool {
		a := conditions[i]
		b := conditions[j]
		return cmp.Less(a.Name, b.Name) || (a.Name == b.Name && cmp.Less(a.RemovalVerb, b.RemovalVerb))
	})

	// Execute template
	b := &bytes.Buffer{}
	tmplContext := map[string]interface{}{"PackageName": packageName, "ObjectType": object.Name, "Conditions": conditions}
	if err := tmpl.ExecuteTemplate(b, "zz_generated.OBJECT.go.tmpl", tmplContext); err != nil {
		return fmt.Errorf("failed to generate '%s': %w", genFilename, err)
	}

	// Format results for Go standards - so IDEs won't complain/change the code
	if formattedBytes, err := format.Source(b.Bytes()); err != nil {
		return fmt.Errorf("failed to format '%s': %w", genFilename, err)
	} else if err := os.WriteFile(genFilename, formattedBytes, 0644); err != nil {
		return fmt.Errorf("failed to write to '%s': %w", genFilename, err)
	}
	return nil
}

func generateConstantsFile(tmpl *template.Template, dir string, packageName string, constants map[string]string) error {
	genFilename := fmt.Sprintf("%s/zz_generated.constants.go", filepath.Dir(dir))

	// Sort data to prevent diffs due to unpredictable symbol parsing order
	var constantsList [][]string
	for k, v := range constants {
		constantsList = append(constantsList, []string{k, v})
	}
	sort.Slice(constantsList, func(i, j int) bool { return cmp.Less(constantsList[i][0], constantsList[j][0]) })

	// Execute template
	b := &bytes.Buffer{}
	tmplContext := map[string]interface{}{"PackageName": packageName, "Constants": constantsList}
	if err := tmpl.ExecuteTemplate(b, "zz_generated.conditions.go.tmpl", tmplContext); err != nil {
		return fmt.Errorf("failed to generate '%s': %w", genFilename, err)
	}

	// Format results for Go standards - so IDEs won't complain/change the code
	if formattedBytes, err := format.Source(b.Bytes()); err != nil {
		return fmt.Errorf("failed to format '%s': %w", genFilename, err)
	} else if err := os.WriteFile(genFilename, formattedBytes, 0644); err != nil {
		return fmt.Errorf("failed to write to '%s': %w", genFilename, err)
	}
	return nil
}

func main() {
	// TODO: replace `go-flags` with `arikkfir/command`

	tmpl, err := template.ParseFS(templatesFS, "*.tmpl")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed parsing embedded Go template: %v\n", err)
		os.Exit(1)
	}

	cfg := Config{}
	configParser := flags.NewParser(&cfg, flags.HelpFlag|flags.PassDoubleDash)
	if _, err := configParser.Parse(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", err)
		configParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	files, err := locateFiles(cfg.Args.Paths)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	constantsByDir := map[string]*Constants{}
	fileSet := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fileSet, file, nil, parser.ParseComments)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed parsing '%s': %v\n", file, err)
			os.Exit(1)
		}

		// We have to read docs of types via the "doc" package, because of how Go parses comments. The "doc" package
		// parses comments in a way that is compatible with the "go/doc" package, which is used by the "kubebuilder"
		// package to parse comments.
		p, err := doc.NewFromFiles(fileSet, []*ast.File{f}, "./")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed parsing '%s': %v\n", file, err)
			os.Exit(1)
		}
		docs := map[string]*doc.Type{}
		for _, t := range p.Types {
			docs[t.Name] = t
		}

		for _, object := range f.Scope.Objects {
			if object.Kind == ast.Typ {
				if typeSpec, ok := object.Decl.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						if docType, ok := docs[object.Name]; ok {
							if comment := docType.Doc; len(comment) > 0 {
								lines := strings.Split(comment, "\n")
								crd := false
								for _, line := range lines {
									if kubeBuilderObjectRootRegexp.MatchString(line) {
										crd = true
										break
									}
								}
								if crd {
									conditionTypes, err := parseObjectConditionTypes(object, lines)
									if err != nil {
										_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
										os.Exit(1)
									}
									if len(conditionTypes) > 0 {
										if err := generateConditionsFile(tmpl, file, f.Name.Name, object, conditionTypes); err != nil {
											_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
											os.Exit(1)
										}
										for _, condition := range conditionTypes {
											constants, ok := constantsByDir[filepath.Dir(file)]
											if !ok {
												constants = &Constants{
													PackageName: f.Name.Name,
												}
												constantsByDir[filepath.Dir(file)] = constants
											}
											constants.Conditions = append(constants.Conditions, condition)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	for dir, constants := range constantsByDir {
		mergedConstants := make(map[string]string)
		for _, condition := range constants.Conditions {
			if _, ok := mergedConstants[condition.Name]; !ok {
				mergedConstants[condition.Name] = condition.Name
			}
			if _, ok := mergedConstants[condition.RemovalVerb]; !ok {
				mergedConstants[condition.RemovalVerb] = condition.RemovalVerb
			}
			for _, reason := range condition.Reasons {
				isCommonReason := false
				if _, ok := mergedConstants[reason]; !ok {
					for _, info := range k8s.CommonConditionReasons {
						for _, commonReason := range info.Reasons {
							if commonReason == reason {
								isCommonReason = true
								break
							}
						}
					}
					if !isCommonReason {
						mergedConstants[reason] = reason
					}
				}
			}
		}
		if err := generateConstantsFile(tmpl, dir, constants.PackageName, mergedConstants); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
}
