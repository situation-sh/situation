package main

import (
	"flag"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	dir        string
	outdir     string
	importPath string
)

func init() {
	// Log to console
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	var err error
	flag.StringVar(&dir, "d", ".", "directory of a Go package")
	flag.StringVar(&outdir, "o", "/tmp", "output directory of markdown files")
	flag.Parse()

	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	p, fset, files := parseSourceDirectory(dir)

	modules := parseModules(p, fset)

	modules = getModuleImports(modules, fset, files)

	log.Infof("Modules found: %d", len(modules))
	for _, m := range modules {
		f := path.Join(outdir, strings.ReplaceAll(m.SrcFile, ".go", ".md"))
		mkdocs := m.MkDocs()
		log.Infof("Writing docs of %s to %s", m.Name, f)
		err := os.WriteFile(f, mkdocs, 0600)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = os.WriteFile(
		path.Join(outdir, "index.md"),
		buildIndex(modules),
		0600,
	)
	if err != nil {
		log.Fatal(err)
	}

}
