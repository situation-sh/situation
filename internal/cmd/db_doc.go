package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/situation-sh/situation/pkg/store"
	"github.com/uptrace/bun/schema"
	"github.com/urfave/cli/v3"
)

const (
	DB_KEY_ICON = "+mynaui:key+"
	DB_PK_ICON  = "+mynaui:link-one+"
)

var (
	dbDocsOutputDir string
)

var legend = strings.Join(
	[]string{
		fmt.Sprintf("%s Primary Key", DB_PK_ICON),
		fmt.Sprintf("%s Foreign Key", DB_KEY_ICON),
		fmt.Sprintf("%s %s Unique", uniqueIcon(1), uniqueIcon(2)),
	},
	"&emsp;",
)

const dbDocDescription = `Generate markdown file describing database tables.
It generates sqlite.md and postgres.md.`

var dbDocCmd = cli.Command{
	Name:        "db-doc",
	Usage:       "Generate database documentation",
	Description: dbDocDescription,
	Action:      dbDocAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "no-sqlite",
			Value:       false,
			Destination: &noSqlite,
		},
		&cli.BoolFlag{
			Name:        "no-postgres",
			Value:       false,
			Destination: &noPostgres,
		},
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Value:       path.Join(rootDir(), "docs", "database"),
			Destination: &dbDocsOutputDir,
		},
	},
}

func dbWrite(storage *store.BunStorage, filename string, title string) error {
	content := [][]byte{
		[]byte("---"),
		fmt.Appendf(nil, "title: %s", title),
		fmt.Appendf(nil, "summary: Database schema for %s storage", title),
		[]byte("---"),
		[]byte(legend),
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	for _, model := range store.TrackedModels {
		table := storage.DB().Table(reflect.TypeOf(model))
		fields := parseTable(table)
		content = append(content, fmt.Appendf(nil, "## %s\n", table.Name))
		content = append(content, markdownify(fields))
	}
	if _, err := file.Write(bytes.Join(content, []byte("\n\n"))); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func dbDocAction(ctx context.Context, cmd *cli.Command) error {
	if !noSqlite {
		storage, err := store.NewSQLiteBunStorage(":memory:", "test-agent", func(err error) {
			logger.WithField("module", "storage").WithError(err).Error("Storage error")
		})
		if err != nil {
			return err
		}
		if err := dbWrite(storage, path.Join(dbDocsOutputDir, "sqlite.md"), "SQLite"); err != nil {
			return err
		}
	}

	if !noPostgres {
		storage, err := store.NewPostgresBunStorageNoPing("postgres://user:pass@localhost:5432/dbname?sslmode=disable", "test-agent", func(err error) {
			logger.WithField("module", "storage").WithError(err).Error("Storage error")
		})
		if err != nil {
			return err
		}
		if err := dbWrite(storage, path.Join(dbDocsOutputDir, "postgres.md"), "PostgreSQL"); err != nil {
			return err
		}
	}
	return nil
}

type FK struct {
	TableName string
	FieldName string
}

type TableField struct {
	Name          string
	Type          string
	IsPK          bool
	NotNull       bool
	AutoIncrement bool
	FK            *FK
	Unique        int
}

func parseTable(table *schema.Table) []*TableField {
	unique := map[string]int{}
	k := 1
	for _, uniqueGroup := range table.Unique {
		for _, field := range uniqueGroup {
			unique[field.Name] = k
		}
		k++
	}

	relations := make(map[string]*FK)
	for _, relation := range table.Relations {
		local := relation.BasePKs[0]
		foreign := relation.JoinPKs[0]
		if local.Name == "id" {
			// seems to be a reverse relation, we want to ignore it
			continue
		}
		relations[local.Name] = &FK{
			TableName: relation.JoinTable.Name,
			FieldName: foreign.Name,
		}
	}

	// fmt.Printf("%s | Relations: %#+v\n", table.Name, table.Relations)
	fields := make([]*TableField, 0, len(table.Fields))
	for _, field := range table.Fields {
		tf := TableField{
			Name:          field.Name,
			Type:          strings.ToUpper(field.DiscoveredSQLType),
			IsPK:          field.IsPK,
			NotNull:       field.NotNull,
			AutoIncrement: field.AutoIncrement,
		}
		if u, ok := unique[field.Name]; ok {
			tf.Unique = u
		}
		if fk, ok := relations[field.Name]; ok {
			tf.FK = fk
		}
		fields = append(fields, &tf)
	}
	return fields
}

func markdownify(fields []*TableField) []byte {
	var md strings.Builder
	md.WriteString("| Name | Type |  |\n")
	md.WriteString("|------|------|-------------|\n")
	for _, field := range fields {
		extra := []string{}
		if field.IsPK {
			extra = append(extra, DB_PK_ICON)
		}
		if field.Unique > 0 {
			extra = append(extra, uniqueIcon(field.Unique))
		}
		if field.FK != nil {
			extra = append(extra, fmt.Sprintf("[%s](#%s)", DB_KEY_ICON, field.FK.TableName))
		}
		md.WriteString(fmt.Sprintf("| `%v` | `%v` | %v |\n", field.Name, field.Type, strings.Join(extra, " ")))
	}
	return []byte(md.String())
}

func uniqueIcon(unique int) string {
	switch unique {
	case 1:
		return "+mynaui:one-diamond-solid+"
	case 2:
		return "+mynaui:two-diamond-solid+"
	case 3:
		return "+mynaui:three-diamond-solid+"
	default:
		return ""
	}
}
