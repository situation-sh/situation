package cmd

import (
	"context"
	"fmt"
	"testing"
)

func TestCmd(t *testing.T) {
	// run all subcommands (not run)
	// special context while testing
	ctx := context.WithValue(context.Background(), "testing", true)
	for _, sub := range app.Commands {
		fmt.Println("RUN>", app.Name, sub.Name)
		if err := app.Run(ctx, []string{app.Name, sub.Name}); err != nil {
			switch err.(type) {
			case *updateUnavailableError:
				// ignore this error
				t.Log(err)
			default:
				t.Error(err)
			}
		}
	}

	if err := app.Run(ctx, []string{app.Name, "--scans", "1"}); err != nil {
		t.Log(err)
	}
}
