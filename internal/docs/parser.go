package docs

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

	"github.com/sirupsen/logrus"
)

type Parser struct {
	ModulesDir string
	Modules    []*ModuleDoc
	logger     logrus.FieldLogger
}

func NewParser(modulesDir string, logger logrus.FieldLogger) *Parser {
	return &Parser{
		ModulesDir: modulesDir,
		logger:     logger,
		Modules:    make([]*ModuleDoc, 0),
	}
}

func (p *Parser) Parse() error {
	pkg, fset, files, err := parseSourceDirectory(p.ModulesDir)
	if err != nil {
		return err
	}
	if err := p.parseModules(pkg, fset); err != nil {
		return err
	}
	p.getModuleImports(fset, files)
	return nil
}

func collectFiles(directory string) (*token.FileSet, []*ast.File, error) {
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
			continue
		}
		f := path.Join(directory, info.Name())

		if strings.HasSuffix(f, ".go") {
			src, err := os.ReadFile(f) //#nosec G304 -- If the file is not valid, that will be triggered by parser.ParseFile
			if err != nil {
				continue
			}
			// parse
			af, err := parser.ParseFile(fset, info.Name(), string(src), parser.ParseComments)
			if err != nil {
				continue
			}
			files = append(files, af)
		}
	}
	return fset, files, nil
}

func getModuleTypes(p *doc.Package) []*doc.Type {
	out := make([]*doc.Type, 0)
	for _, t := range p.Types {
		if strings.HasSuffix(t.Name, "Module") && len(t.Name) > 6 {
			out = append(out, t)
		}
	}
	return out
}

func getModuleTypeName(t *doc.Type) (string, error) {
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

func (parser *Parser) parseModules(p *doc.Package, fset *token.FileSet) error {
	internal := make(map[string]*ModuleDoc)
	// loop over the modules
	for _, t := range getModuleTypes(p) {
		parser.logger.WithField("object", t.Name).Info("Get module name")
		// Get module name of %s object
		name, err := getModuleTypeName(t)
		if err != nil {
			parser.logger.
				WithError(err).
				WithField("module-name", t.Name).
				Warn("Fail to parse module. Skipping.")
			continue
		}

		m := NewModuleDoc(parser.ModulesDir)
		m.Name = name
		m.Object = t
		m.Synopsis = p.Synopsis(t.Doc)
		m.RawMarkdown = p.Markdown(t.Doc)
		m.SrcFile = fset.Position(t.Decl.TokPos).Filename

		parser.logger.WithField("module-name", t.Name).Infof("Get module dependencies")
		deps, err := getModuleTypeDependencies(t)
		if err != nil {
			return err
			// parser.logger.
			// 	WithError(err).
			// 	WithField("module", t.Name).
			// 	Fatal("Fail to parse module dependencies. Skipping.")

			// continue
		}
		m.Dependencies = deps

		internal[t.Name] = m
	}

	parser.logger.Info("Populating modules status")
	for _, key := range []string{"LINUX", "WINDOWS", "MACOS", "ROOT"} {
		for _, note := range p.Notes[key] {
			m := internal[note.UID]
			m.SetStatus(key, note.Body)
		}
	}

	for _, m := range internal {
		parser.Modules = append(parser.Modules, m)
	}

	return nil
}

func (parser *Parser) getModuleImports(fset *token.FileSet, files []*ast.File) {
	parser.logger.Info("Populating modules imports")
	map1 := make(map[string][]string)

	for _, f := range files {
		name := fset.Position(f.Package).Filename
		map1[name] = make([]string, len(f.Imports))
		for k, imp := range f.Imports {
			map1[name][k] = strings.ReplaceAll(imp.Path.Value, "\"", "")
		}
	}

	for _, m := range parser.Modules {
		m.Imports = map1[m.SrcFile]
		parser.logger.
			WithField("module-name", m.Name).
			Info("Get module imports")
	}
}

// parseSourceDirectory return the package doc along with the source files through a
// fileset and the file AST
func parseSourceDirectory(directory string) (*doc.Package, *token.FileSet, []*ast.File, error) {
	ip, err := detectImportPath(directory, "", 3)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while detecting import path: %v", err)
	}
	fset, files, err := collectFiles(directory)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while collecting files: %v", err)
	}
	// PreserveAST to access function body notably
	p, err := doc.NewFromFiles(fset, files, ip, doc.PreserveAST)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while creating package doc: %v", err)
	}
	return p, fset, files, nil
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

const LEGEND = `
<div style="display: flex; flex-direction: row; gap: 1.5rem; font-family: monospace; align-items: center;">
	<div style="display: flex; flex-direction: row; gap: 0.5rem; align-items: center;">
		<img src="{{ linux_icon_src }}" alt="linux" />
		<span>Linux</span>
	</div>
	<div style="display: flex; flex-direction: row; gap: 0.5rem; align-items: center;">
		<img src="{{ windows_icon_src }}" alt="windows" />
		<span>Windows</span>
	</div>
	<div style="display: flex; flex-direction: row; gap: 0.5rem; align-items: center;">
		<img src="{{ root_required_icon_src }}" alt="root-required" />
		<span>Root required</span>
	</div>
</div>
`

func (parser *Parser) BuildIndex() []byte {
	head := "---\ntitle: Modules reference\nsummary: List of all collectors\nsidebar_title: Reference\n---\n\n"
	head += LEGEND
	head += "| Name | Summary | Dependencies | Status |\n"
	head += "|------|---------|--------------|--------|"
	line := "\n| %s   | %s      | %s           | %s     |"

	out := []byte(head)
	links := make(map[string]string)
	for _, m := range parser.Modules {
		links[m.Name] = fmt.Sprintf(`[%s](%s)`, m.Name, strings.Replace(m.SrcFile, ".go", ".md", 1))
	}
	// sort modules by name
	sort.Slice(parser.Modules, func(i, j int) bool {
		return parser.Modules[i].Name < parser.Modules[j].Name
	})
	for _, m := range parser.Modules {
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
