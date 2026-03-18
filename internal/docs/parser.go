package docs

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
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
	if err := p.parseModules(pkg, fset, files); err != nil {
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
					s := strings.ReplaceAll(basic.Value, "\"", "")
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
								out = append(out, strings.ReplaceAll(basic.Value, "\"", ""))
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

func (parser *Parser) parseModules(p *doc.Package, fset *token.FileSet, _ []*ast.File) error {
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
		m.Options = getModuleTypeOptions(t, fset, parser.ModulesDir)

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

// exprString renders an AST expression back to its source representation.
func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return ""
	}
	return buf.String()
}

// extractSetDefaultCall returns the *ast.CallExpr for setDefault(...) within a
// statement, handling the three forms used in practice:
//
//	if err := setDefault(...); err != nil { ... }
//	return setDefault(...)
//	setDefault(...)
func extractSetDefaultCall(stmt ast.Stmt) *ast.CallExpr {
	var ce *ast.CallExpr
	switch s := stmt.(type) {
	case *ast.IfStmt:
		if assign, ok := s.Init.(*ast.AssignStmt); ok && len(assign.Rhs) == 1 {
			ce, _ = assign.Rhs[0].(*ast.CallExpr)
		}
	case *ast.ReturnStmt:
		if len(s.Results) == 1 {
			ce, _ = s.Results[0].(*ast.CallExpr)
		}
	case *ast.ExprStmt:
		ce, _ = s.X.(*ast.CallExpr)
	}
	if ce == nil {
		return nil
	}
	ident, ok := ce.Fun.(*ast.Ident)
	if !ok || ident.Name != "setDefault" {
		return nil
	}
	return ce
}

// findCompositeLit returns the *ast.CompositeLit for the given struct type name
// within an expression, unwrapping & if present.
func findCompositeLit(expr ast.Expr, typeName string) *ast.CompositeLit {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return findCompositeLit(e.X, typeName)
		}
	case *ast.CompositeLit:
		if ident, ok := e.Type.(*ast.Ident); ok && ident.Name == typeName {
			return e
		}
	}
	return nil
}

// extractInitDefaults re-parses all .go files in modulesDir independently
// (avoiding any doc.NewFromFiles side-effects) and returns a map of struct
// field name → default value string for the given structName.
func extractInitDefaults(modulesDir, structName string) map[string]string {
	defaults := make(map[string]string)

	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return defaults
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		src, err := os.ReadFile(path.Join(modulesDir, entry.Name())) //#nosec G304 -- path is limited to modulesDir
		if err != nil {
			continue
		}
		f, err := parser.ParseFile(fset, entry.Name(), string(src), 0)
		if err != nil {
			continue
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "init" || fn.Body == nil {
				continue
			}

			// First pass: collect variable assignments  m := &TypeName{...}
			varLits := make(map[string]*ast.CompositeLit)
			for _, stmt := range fn.Body.List {
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok || len(assign.Rhs) != 1 {
					continue
				}
				lit := findCompositeLit(assign.Rhs[0], structName)
				if lit == nil {
					continue
				}
				for _, lhs := range assign.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						varLits[ident.Name] = lit
					}
				}
			}

			// Second pass: find registerModule calls
			for _, stmt := range fn.Body.List {
				exprStmt, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}
				ce, ok := exprStmt.X.(*ast.CallExpr)
				if !ok {
					continue
				}
				ident, ok := ce.Fun.(*ast.Ident)
				if !ok || ident.Name != "registerModule" || len(ce.Args) == 0 {
					continue
				}

				// Direct composite literal: registerModule(&TypeName{...})
				lit := findCompositeLit(ce.Args[0], structName)
				// Variable reference: registerModule(m)
				if lit == nil {
					if argIdent, ok := ce.Args[0].(*ast.Ident); ok {
						lit = varLits[argIdent.Name]
					}
				}
				if lit == nil {
					continue
				}

				for _, elt := range lit.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := kv.Key.(*ast.Ident)
					if !ok {
						continue
					}
					defaults[key.Name] = exprString(fset, kv.Value)
				}
				return defaults // found the registerModule call for this type
			}
		}
	}
	return defaults
}

// getModuleTypeOptions extracts the options declared in the Bind() method of a
// module type, returning the option name, its Go type, and its default value.
func getModuleTypeOptions(t *doc.Type, fset *token.FileSet, modulesDir string) []ModuleOption {
	typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil
	}

	// Build field name → Go type string from the struct declaration
	fieldTypes := make(map[string]string)
	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		for _, field := range structType.Fields.List {
			typeName := exprString(fset, field.Type)
			for _, name := range field.Names {
				fieldTypes[name.Name] = typeName
			}
		}
	}

	// Extract option key → field name pairs from Bind()
	type pair struct{ key, field string }
	var pairs []pair
	for _, method := range t.Methods {
		if method.Name != "Bind" {
			continue
		}
		for _, stmt := range method.Decl.Body.List {
			ce := extractSetDefaultCall(stmt)
			if ce == nil || len(ce.Args) < 5 {
				continue
			}
			keyLit, ok := ce.Args[2].(*ast.BasicLit)
			if !ok || keyLit.Kind != token.STRING {
				continue
			}
			key := strings.Trim(keyLit.Value, `"`)

			fieldName := ""
			if unary, ok := ce.Args[3].(*ast.UnaryExpr); ok && unary.Op == token.AND {
				if sel, ok := unary.X.(*ast.SelectorExpr); ok {
					fieldName = sel.Sel.Name
				}
			}
			pairs = append(pairs, pair{key, fieldName})
		}
	}

	if len(pairs) == 0 {
		return nil
	}

	defaults := extractInitDefaults(modulesDir, typeSpec.Name.Name)

	options := make([]ModuleOption, 0, len(pairs))
	for _, p := range pairs {
		options = append(options, ModuleOption{
			Name:    p.key,
			Type:    fieldTypes[p.field],
			Default: defaults[p.field],
		})
	}
	return options
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
