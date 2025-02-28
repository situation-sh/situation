package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func collectFiles(directory string) (*token.FileSet, []*ast.File, error) {
	log.Debugf("Collecting files from %s", directory)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	files := make([]*ast.File, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Warnf("ignoring %s (cannot read entry info: %v)", entry.Name(), err)
			continue
		}
		f := path.Join(directory, info.Name())

		if strings.HasSuffix(f, ".go") {
			src, err := os.ReadFile(f) //#nosec G304 -- If the file is not valid, that will be triggered by parser.ParseFile
			if err != nil {
				log.Warnf("ignoring %s (cannot read file: %v)", entry.Name(), err)
				continue
			}
			// parse
			af, err := parser.ParseFile(fset, info.Name(), string(src), parser.ParseComments)
			if err != nil {
				log.Warnf("ignoring %s (cannot parse file: %v)", entry.Name(), err)
				continue
			}
			files = append(files, af)
		}
	}
	return fset, files, nil
}

func getModuleTypes(p *doc.Package) []*doc.Type {
	log.Infof("Get modules from package '%s'", p.Name)
	out := make([]*doc.Type, 0)
	for _, t := range p.Types {
		if strings.HasSuffix(t.Name, "Module") && len(t.Name) > 6 {
			out = append(out, t)
		}
	}
	return out
}

func getModuleTypeName(t *doc.Type) (string, error) {
	log.Infof("Get module name of %s object", t.Name)
	for _, method := range t.Methods {
		// find the Name() method
		if method.Name != "Name" {
			continue
		}
		// loop over internal statements to find the 'Return'
		for _, smt := range method.Decl.Body.List {
			if returnSmt, ok := smt.(*ast.ReturnStmt); ok {
				// the return statement is assumed to be a literal of basic type
				expr := returnSmt.Results[0]
				if basic, ok := expr.(*ast.BasicLit); ok {
					// remove wrapping quotes
					s := strings.Replace(basic.Value, "\"", "", -1)
					return s, nil
				}
			}
		}
	}
	return "", fmt.Errorf("cannot get module name of %s object", t.Name)
}

func getModuleTypeDependencies(t *doc.Type) ([]string, error) {
	log.Infof("Get module dependencies of %s object", t.Name)
	out := make([]string, 0)

	for _, method := range t.Methods {
		// find the Name() method
		if method.Name != "Dependencies" {
			continue
		}
		// loop over internal statements to find the 'Return'
		for _, smt := range method.Decl.Body.List {
			if returnSmt, ok := smt.(*ast.ReturnStmt); ok {
				// the return statement is assumed to be a composite of
				// literal of basic types
				expr := returnSmt.Results[0]
				if composite, ok := expr.(*ast.CompositeLit); ok {
					// when the list is empty: composite.Elts == nil
					if composite.Elts != nil {
						for _, element := range composite.Elts {
							if basic, ok := element.(*ast.BasicLit); ok {
								// remove wrapping quotes
								out = append(out, strings.Replace(basic.Value, "\"", "", -1))
							}
						}
					}
					return out, nil
				}
				if ident, ok := expr.(*ast.Ident); ok {
					if ident.Name == "nil" {
						return []string{}, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("cannot get module dependencies of %s object", t.Name)
}

func parseModules(p *doc.Package, fset *token.FileSet) []*ModuleDoc {
	internal := make(map[string]*ModuleDoc)
	// loop over the modules
	for _, t := range getModuleTypes(p) {
		name, err := getModuleTypeName(t)
		if err != nil {
			log.Fatalf("error while parsing module %s: %v", t.Name, err)
		}

		m := NewModuleDoc()
		m.Name = name
		m.Object = t
		m.Synopsis = p.Synopsis(t.Doc)
		m.RawMarkdown = p.Markdown(t.Doc)
		m.SrcFile = fset.Position(t.Decl.TokPos).Filename

		deps, err := getModuleTypeDependencies(t)
		if err != nil {
			log.Fatalf("error while parsing module %s: %v", t.Name, err)
		}
		m.Dependencies = deps

		internal[t.Name] = m
	}

	log.Info("Populating modules status")
	for _, key := range []string{"LINUX", "WINDOWS", "MACOS", "ROOT"} {
		for _, note := range p.Notes[key] {
			m := internal[note.UID]
			m.SetStatus(key, note.Body)
		}
	}

	out := make([]*ModuleDoc, 0)
	for _, m := range internal {
		out = append(out, m)
	}
	return out
}

func getModuleImports(mods []*ModuleDoc, fset *token.FileSet, files []*ast.File) []*ModuleDoc {
	log.Info("Populating modules imports")
	map1 := make(map[string][]string)

	for _, f := range files {
		name := fset.Position(f.Package).Filename
		map1[name] = make([]string, len(f.Imports))
		for k, imp := range f.Imports {
			map1[name][k] = strings.ReplaceAll(imp.Path.Value, "\"", "")
		}
	}

	for _, m := range mods {
		m.Imports = map1[m.SrcFile]
		log.Debugf("Get module imports for %s", m.Name)
	}

	return mods
}

// parseSourceDirectory return the package doc along with the source files through a
// fileset and the file AST
func parseSourceDirectory(directory string) (*doc.Package, *token.FileSet, []*ast.File) {
	log.Infof("Parsing sources within %s", directory)
	ip, err := detectImportPath(dir, "", 3)
	if err != nil {
		log.Fatalf("error while detecting import path: %v", err)
	}
	fset, files, err := collectFiles(directory)
	if err != nil {
		log.Fatalf("error while collecting files: %v", err)
	}
	// PreserveAST to access function body notably
	p, err := doc.NewFromFiles(fset, files, ip, doc.PreserveAST)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Package name: %s, Files: %d", p.Name, len(p.Filenames))
	return p, fset, files
}

func buildBadge(m *ModuleDoc) string {
	badges := make([]string, 0)
	if m.Status.LINUX == YES {
		badges = append(badges, "{{ linux_ok }}")
	}
	if m.Status.WINDOWS == YES {
		badges = append(badges, "{{ windows_ok }}")
	}
	if m.Status.MACOS == YES {
		badges = append(badges, "{{ macos_ok }}")
	}
	if m.Status.ROOT == YES {
		badges = append(badges, "{{ root_required }}")
	}
	return strings.Join(badges, " ")
}

func buildIndex(modules []*ModuleDoc) []byte {
	head := "| Name | Summary | Dependencies | Status |\n"
	head += "|------|---------|--------------|--------|"
	line := "\n| %s   | %s      | %s           | %s     |"

	out := []byte(head)
	links := make(map[string]string)
	for _, m := range modules {
		links[m.Name] = fmt.Sprintf("[%s](%s)", m.Name, strings.Replace(m.SrcFile, ".go", ".md", 1))
	}
	// sort modules by name
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})
	for _, m := range modules {
		// deps with links
		deps := make([]string, len(m.Dependencies))
		for i, d := range m.Dependencies {
			deps[i] = links[d]
		}
		// badges
		badge := buildBadge(m)
		fmtline := fmt.Sprintf(line, links[m.Name], m.Synopsis, strings.Join(deps, ", "), badge)
		out = append(out, []byte(fmtline)...)
	}

	return out
}
