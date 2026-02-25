package cmd

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/situation-sh/situation/internal/docs"
	"github.com/urfave/cli/v3"
)

var (
	modulesDir          string
	moduleDocsOutputDir string
)

const ModulesDocDescription = `This code aims to generate markdown files from module source files.
Basically, if DoStuffModule is defined in modules/do_stuff.go, 
the tool generates docs/modules/do_stuff.md. 
We cannot select  a single module: the tool treats all the modules 
at once and also updates docs/modules/index.md.`

var modulesDocCmd = cli.Command{
	Name:        "modules-doc",
	Usage:       "Generate modules documentation",
	Description: ModulesDocDescription,
	Action:      modulesDocAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Value:       path.Join(rootDir(), "docs", "modules"),
			Destination: &moduleDocsOutputDir,
		},
		&cli.StringFlag{
			Name:        "modules-dir",
			Aliases:     []string{"d"},
			Value:       path.Join(rootDir(), "pkg", "modules"),
			Destination: &modulesDir,
		},
	},
}

func modulesDocAction(ctx context.Context, cmd *cli.Command) error {
	parser := docs.NewParser(modulesDir, logger.WithField("module", "parser"))
	if err := parser.Parse(); err != nil {
		return err
	}
	logger.WithField("modules", len(parser.Modules)).Info("Modules found")

	for _, m := range parser.Modules {
		f := path.Join(moduleDocsOutputDir, strings.ReplaceAll(m.SrcFile, ".go", ".md"))
		mkdocs := m.MkDocs()
		logger.WithField("name", m.Name).WithField("file", f).Info("Writing module docs")
		if err := os.WriteFile(f, mkdocs, 0644); err != nil {
			return err
		}
	}

	indexFile := path.Join(moduleDocsOutputDir, "index.md")
	logger.
		WithField("name", "index").
		WithField("file", indexFile).
		Info("Writing module index")
	err := os.WriteFile(
		indexFile,
		parser.BuildIndex(),
		0644,
	)
	return err
}
