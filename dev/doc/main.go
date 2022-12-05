package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	dir        string
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

func detectImportPath(dir string, base string, levels int) (string, error) {
	d := filepath.Clean(dir)

	if levels < 0 {
		return "", fmt.Errorf("directory depth exceeded")
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "go.mod" {
			file, err := os.Open(path.Join(dir, entry.Name()))
			if err != nil {
				return "", err
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "module") {
					words := strings.SplitN(line, " ", 2)
					return filepath.Join(words[1], base), nil
				}
			}
		}
	}
	// no entry found
	b := filepath.Base(d)
	return detectImportPath(filepath.Dir(d), filepath.Join(b, base), levels-1)
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
			log.Debugf("ignoring %s (directory)", entry.Name())
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Warn("ignoring %s (cannot read entry info: %v)", entry.Name(), err)
			continue
		}
		f := path.Join(directory, info.Name())

		if strings.HasSuffix(f, ".go") {
			src, err := os.ReadFile(f)
			if err != nil {
				log.Warn("ignoring %s (cannot read file: %v)", entry.Name(), err)
				continue
			}
			// parse
			af, err := parser.ParseFile(fset, info.Name(), string(src), parser.ParseComments)
			if err != nil {
				log.Warn("ignoring %s (cannot parse file: %v)", entry.Name(), err)
				continue
			}
			files = append(files, af)
		}
	}
	return fset, files, nil
}

func main() {
	var err error
	flag.StringVar(&dir, "d", ".", "directory of a Go package")
	flag.StringVar(&importPath, "i", "", "import path of the package (by default it tries to build it)")
	flag.Parse()

	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	if importPath == "" {
		importPath, err = detectImportPath(dir, "", 3)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Infof("Import path: %s", importPath)

	fset, files, err := collectFiles(dir)
	if err != nil {
		log.Fatal(err)
	}

	p, err := doc.NewFromFiles(fset, files, importPath)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(p.Notes)
	// for _, t := range p.Types {
	// 	fmt.Printf("%s %s\n", t.Name, t.Doc)
	// }

	s := NewFromPackageDoc(p)
	fmt.Println(s)
}
