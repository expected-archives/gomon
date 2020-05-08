package imports

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func FromFile(repo string, filepath string, result map[string]bool) {
	for _, importPath := range getImports(filepath) {
		if strings.HasPrefix(importPath, repo) {
			pkgPath := strings.Replace(importPath, repo+"/", "", -1)
			if _, alreadyChecked := result[pkgPath]; !alreadyChecked {
				// add in cache this pkg
				result[pkgPath] = true

				files := getFiles(pkgPath)

				// for each files in this pkgPath add to result all imports
				for _, file := range files {
					result[file] = true
					FromFile(repo, file, result)
				}
			}
		}
	}
}

func getImports(filepath string) []string {
	set := token.NewFileSet()
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil
	}
	astFile, err := parser.ParseFile(set, filepath, bytes, 0)
	if err != nil {
		return nil
	}

	var paths []string
	importList := imports(set, astFile)
	for _, list := range importList {
		for _, i := range list {
			paths = append(paths, i.Path.Value[1:len(i.Path.Value)-1])
		}
	}

	return paths
}

func imports(tokenFileSet *token.FileSet, f *ast.File) [][]*ast.ImportSpec {
	var groups [][]*ast.ImportSpec

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			break
		}

		var group []*ast.ImportSpec

		var lastLine int
		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)
			pos := importSpec.Path.ValuePos
			line := tokenFileSet.Position(pos).Line
			if lastLine > 0 && pos > 0 && line-lastLine > 1 {
				groups = append(groups, group)
				group = []*ast.ImportSpec{}
			}
			group = append(group, importSpec)
			lastLine = line
		}
		groups = append(groups, group)
	}

	return groups
}

func getFiles(pkgPath string) (files []string) {
	if err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil
	}

	return files
}
