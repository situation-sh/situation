package config

import (
	"flag"
	"fmt"

	"github.com/asiffer/puzzle"
	"github.com/asiffer/puzzle/flagset"
	"github.com/asiffer/puzzle/jsonfile"
	"github.com/asiffer/puzzle/urfave3"
	"github.com/urfave/cli/v3"
)

var k = puzzle.NewConfig()

func Define[T any](key string, defaultValue T, options ...puzzle.MetadataOption) {
	if err := puzzle.Define(k, key, defaultValue, options...); err != nil {
		panic(err)
	}
}

func DefineVar[T any](key string, boundVariable *T, options ...puzzle.MetadataOption) {
	if err := puzzle.DefineVar(k, key, boundVariable, options...); err != nil {
		panic(err)
	}
}

func DefineVarWithUsage[T any](key string, boundVariable *T, usage string) {
	if err := puzzle.DefineVar(k, key, boundVariable, puzzle.WithDescription(usage)); err != nil {
		panic(err)
	}
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

func PopulateFlags(fs *flag.FlagSet) error {
	return flagset.Populate(k, fs)
}
