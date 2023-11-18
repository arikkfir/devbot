package main

import (
	_ "embed"
	"fmt"
	"github.com/secureworks/errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

const (
	searchGlob = "./*.go"
)

var (
	markerRegExp = regexp.MustCompile(`^\s*//\s*\+StatusCondition:([A-Z][A-Za-z0-9]*):([A-Z][A-Za-z0-9]*)\n?$`)

	//go:embed zz_generated.setstatuscondition.go.tmpl
	goCodeTemplateString string

	goCodeTemplate *template.Template
)

func init() {
	tmpl, err := template.New("generatedCode").Parse(goCodeTemplateString)
	if err != nil {
		fmt.Printf("Failed parsing embedded Go template: %v\n", err)
		os.Exit(1)
	}
	goCodeTemplate = tmpl
}

type TemplateData struct {
	PackageName   string
	ObjectType    string
	ConditionType string
}

func generateFile(tmpl *template.Template, data TemplateData) error {
	generatedFile, err := os.Create(fmt.Sprintf("zz_generated.%s.%s.go", data.ObjectType, data.ConditionType))
	if err != nil {
		return errors.New("failed to create conditions file for '%s.%s(..)'", data.ObjectType, data.ConditionType, err)
	}
	defer generatedFile.Close()

	err = tmpl.Execute(generatedFile, data)
	if err != nil {
		return errors.New("failed to generate Go code file for '%s.%s(..)'", data.ObjectType, data.ConditionType, err)
	}

	return nil
}

func processFile(path string) error {
	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return errors.New("failed parsing '%s'", path, err)
	}

	for _, commentGroup := range f.Comments {
		for _, comment := range commentGroup.List {
			if matches := markerRegExp.FindStringSubmatch(comment.Text); len(matches) == 3 {
				data := TemplateData{
					PackageName:   f.Name.Name,
					ObjectType:    matches[1],
					ConditionType: matches[2],
				}
				if err := generateFile(goCodeTemplate, data); err != nil {
					return errors.New("failed to generate Go code", err)
				}
			}
		}
	}

	return nil
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed getting current working directory: %v", err)
		os.Exit(1)
	}

	matches, err := filepath.Glob(searchGlob)
	if err != nil {
		fmt.Printf("Failed searching files for pattern '%s' in '%s': %v\n", searchGlob, wd, err)
		os.Exit(1)
	}

	for _, match := range matches {
		if err := processFile(match); err != nil {
			fmt.Printf("Failed processing file '%s': %v", match, err)
			os.Exit(1)
		}
	}
}
