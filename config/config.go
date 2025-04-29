package config

import (
	"fmt"

	"github.com/asiffer/puzzle"
	"github.com/asiffer/puzzle/jsonfile"
	"github.com/asiffer/puzzle/urfave3"
	"github.com/urfave/cli/v3"
)

var k = puzzle.NewConfig()

func Define[T any](key string, defaultValue T, options ...puzzle.MetadataOption) error {
	return puzzle.Define(k, key, defaultValue, options...)
}

func DefineVar[T any](key string, boundVariable *T, options ...puzzle.MetadataOption) error {
	return puzzle.DefineVar(k, key, boundVariable, options...)
}

func DefineVarWithUsage[T any](key string, boundVariable *T, usage string) error {
	return puzzle.DefineVar(k, key, boundVariable, puzzle.WithDescription(usage))
}

func Get[T any](key string) (T, error) {
	return puzzle.Get[T](k, key)
}

func Set(key string, value string) error {
	entry, exists := k.GetEntry(key)
	if !exists {
		return fmt.Errorf("key '%s' not found", key)
	}
	return entry.Set(value)
}

func Flags() []cli.Flag {
	flags, err := urfave3.Build(k)
	if err != nil {
		panic(err)
	}
	return flags
}

func JSON() []byte {
	bytes, err := jsonfile.ToJSON(k)
	if err != nil {
		panic(err)
	}
	return bytes
}
