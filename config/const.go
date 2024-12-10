package config

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

// var LogLevel uint = 5

// var LogFile string

// var Scans = 1

// var Period = 2 * time.Minute
