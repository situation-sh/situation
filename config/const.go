package config

import "encoding/hex"

// Version of the agent (it is set during compilation)
var Version = "X.X.X"

// Commit is the short commit hash of the version of
// the agent (it is set during compilation)
var Commit = "SHORTCOMMITHASH"

// ID is the unique identifier of the agent instance
// It is a 32 hexchars (16 bytes) random string
var ID = [...]byte{
	202, 254, 202, 254, 202, 254, 202, 254,
	202, 254, 202, 254, 202, 254, 202, 254,
}

const defaultIDHexString = "cafecafecafecafecafecafecafecafe"

func GetDefaultID() []byte {
	id, err := hex.DecodeString(defaultIDHexString)
	if err != nil {
		panic(err)
	}
	return id
}
