package config

import (
	"encoding/hex"

	"github.com/asiffer/puzzle"
	"github.com/asiffer/puzzle/jsonfile"
	"github.com/asiffer/puzzle/urfave3"
	"github.com/urfave/cli/v3"
)

var (
	Agent   [16]byte = DefaultAgent()
	Version          = "0.0.0"
	Commit           = ""
	Module           = ""
)

func DefaultAgent() [16]byte {
	// sha256(situation)[:16]
	return [16]byte{
		252, 9, 126, 101,
		80, 60, 179, 173,
		158, 184, 225, 15,
		90, 97, 118, 17,
	}
}

func AgentString() string {
	return hex.EncodeToString(Agent[:])
}

var k = puzzle.NewConfig()

type Configurable interface {
	Bind(config *puzzle.Config) error
}

func Define[T any](key string, defaultValue T, options ...puzzle.MetadataOption) error {
	return puzzle.Define(k, key, defaultValue, options...)
}

func DefineVar[T any](key string, boundVariable *T, options ...puzzle.MetadataOption) error {
	return puzzle.DefineVar(k, key, boundVariable, options...)
}

func Get[T any](key string) (T, error) {
	return puzzle.Get[T](k, key)
}

func Bind(cf Configurable) {
	cf.Bind(k)
}

func Urfave3() ([]cli.Flag, error) {
	return urfave3.Build(k)
}

func JSON() ([]byte, error) {
	return jsonfile.ToJSON(k)
}

func UpdateFromJSON(raw []byte) error {
	return jsonfile.ReadJSONRaw(k, raw)
}
func ReadEnv() error {
	return puzzle.ReadEnv(k)
}
