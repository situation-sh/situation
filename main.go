package main

import (
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
