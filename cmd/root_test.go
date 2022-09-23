package cmd

import (
	"fmt"
	"testing"
)

func TestCmd(t *testing.T) {
	for _, sub := range app.Commands {
		fmt.Println(sub.Name)
		if err := app.Run([]string{app.Name, sub.Name}); err != nil {
			t.Error(err)
		}
	}
}
