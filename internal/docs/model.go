package docs

import (
	"bytes"
	"fmt"
	"go/doc"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Status string

const (
	UNKNOWN = Status("unknown")
	YES     = Status("true")
	NO      = Status("false")
	ALPHA   = Status("alpha")
	BETA    = Status("beta")
)

const librariesTemplate = `
### Dependencies

/// tab | Standard library

{% for i in std_imports %}
- [{{ i }}](https://pkg.go.dev/{{ i }})
{% endfor %}

///

/// tab | External

{% for i in imports %}
- [{{ i }}](https://pkg.go.dev/{{ i }})
{% endfor %}

///
`

const headerTemplate = `
{%% if windows == true %%}{{ windows_ok }}{%% endif %%}
{%% if linux == true %%}{{ linux_ok }}{%% endif %%}
{%% if root == true %%}{{ root_required }}{%% endif %%}

%s

### Details
`

const metadataTemplate = `---
linux: %s
windows: %s
macos: %s
root: %s
title: %s
summary: "%s"
date: %v
filename: %s
%s
---
`

// list of keys we should use to document modules
var keys = []string{"LINUX", "WINDOWS", "MACOS", "ROOT"}

type ModuleStatus struct {
	LINUX   Status
	WINDOWS Status
	MACOS   Status
	ROOT    Status
}

type ModuleDoc struct {
	modulesDir   string
	Name         string
	Dependencies []string
	Status       ModuleStatus
	Synopsis     string
	RawMarkdown  []byte
	SrcFile      string
	Imports      []string
	Object       *doc.Type
}

func NewModuleDoc(modulesDir string) *ModuleDoc {
	return &ModuleDoc{
		modulesDir:   modulesDir,
		Dependencies: make([]string, 0),
		RawMarkdown:  make([]byte, 0),
		Imports:      make([]string, 0),
	}
}

func (m *ModuleDoc) SetStatus(key, value string) {
	value = strings.Trim(value, "\n")

	status := StatusFromString(value)
	// fmt.Println(key, value, status)
	switch key {
	case "LINUX":
		m.Status.LINUX = status
	case "WINDOWS":
		m.Status.WINDOWS = status
	case "MACOS":
		m.Status.MACOS = status
	case "ROOT":
		m.Status.ROOT = status
	}
}

func (m *ModuleDoc) Title() string {
	base := strings.TrimSuffix(m.Object.Name, "Module")

	// title for special module like SaaS
	if matched, err := regexp.MatchString("[A-Z][a-z]+[A-Z]", base); err == nil && matched {
		return base
	}

	re := regexp.MustCompile("[A-Z][a-z]+")
	space := byte(32)

	f := func(x []byte) []byte {
		w := append([]byte{space}, x...)
		w = append(w, 32)
		return w
	}
	z := re.ReplaceAllFunc([]byte(base), f)
	z = bytes.ReplaceAll(z, []byte("  "), []byte(" "))
	return strings.TrimSpace(string(z))
}

func (m *ModuleDoc) Summary() string {
	words := strings.Fields(m.Synopsis)
	if len(words) == 0 {
		return ""
	}
	if strings.Contains(words[0], "Module") {
		words[1] = strings.ToUpper(words[1][:1]) + words[1][1:]
		return strings.Join(words[1:], " ")
	}
	return strings.Join(words, " ")
}

func (m *ModuleDoc) insertSubmoduleImports() {
	newImports := make([]string, 0)
	for _, imp := range m.Imports {
		if strings.Contains(imp, "modules/") {
			// sub-module
			w := strings.Split(imp, "modules/")
			// get last element (name of the submodule)
			name := w[len(w)-1] + "/"
			// append that name to the current directory
			subpkg, _, _, _ := parseSourceDirectory(path.Join(m.modulesDir, name))
			// insert its imports
			newImports = append(newImports, subpkg.Imports...)
		} else {
			newImports = append(newImports, imp)
		}
	}
	m.Imports = unique(newImports)
}

func (m *ModuleDoc) ImportHeader() string {
	h := ""
	std := make([]string, 0)
	ext := make([]string, 0)

	// insert submodules import if they exist
	m.insertSubmoduleImports()

	// integrate submodules dependencies
	for _, imp := range m.Imports {
		if strings.HasPrefix(imp, "github.com/situation-sh/situation") {
			// ignore local imports
			continue
		}
		first := strings.Split(imp, "/")[0]
		if strings.Contains(first, ".") {
			ext = append(ext, imp)
		} else {
			std = append(std, imp)
		}
	}

	sort.Strings(std)
	sort.Strings(ext)

	if len(std) == 0 {
		h += "std_imports: []\n"
	} else {
		h += "std_imports:\n  - " + strings.Join(std, "\n  - ") + "\n"
	}

	if len(ext) == 0 {
		h += "imports: []"
	} else {
		h += "imports:\n  - " + strings.Join(ext, "\n  - ")
	}

	return h
}

func (m *ModuleDoc) Libraries() []byte {
	return []byte(librariesTemplate)
}

func (m *ModuleDoc) Markdown() []byte {
	// remove synopsis
	md := bytes.ReplaceAll(m.RawMarkdown, []byte(m.Synopsis), []byte{})
	// replace tabs (inserted by gopls) by space
	md = bytes.ReplaceAll(md, []byte("\t"), []byte(" "))
	// replace escaped backticks by normal backtics
	md = bytes.ReplaceAll(md, []byte("\\`"), []byte("`"))
	// underscore
	md = bytes.ReplaceAll(md, []byte("\\_"), []byte("_"))
	// stars
	md = bytes.ReplaceAll(md, []byte("\\*"), []byte("*"))
	// brackets
	md = bytes.ReplaceAll(md, []byte("\\("), []byte("("))
	md = bytes.ReplaceAll(md, []byte("\\)"), []byte(")"))
	md = bytes.ReplaceAll(md, []byte("\\["), []byte("["))
	md = bytes.ReplaceAll(md, []byte("\\]"), []byte("]"))

	s := fmt.Sprintf(headerTemplate, m.Synopsis)
	return append([]byte(s), md...)
}

func (m *ModuleDoc) MkDocs() []byte {
	h := fmt.Appendf(nil, metadataTemplate,
		m.Status.LINUX, m.Status.WINDOWS, m.Status.MACOS, m.Status.ROOT,
		m.Title(), m.Summary(), time.Now().Format("2006-01-02"), m.SrcFile,
		m.ImportHeader())

	h = append(h, m.Markdown()...)
	h = append(h, m.Libraries()...)
	return h
}

func StatusFromString(s string) Status {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "yes", "ok":
		return YES
	case "no", "ko", "error":
		return NO
	case "alpha":
		return ALPHA
	case "beta":
		return BETA
	default:
		return UNKNOWN
	}
}
