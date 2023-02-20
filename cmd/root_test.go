package cmd

import (
	"context"
	"fmt"
	"testing"
)

func TestCmd(t *testing.T) {
	// special context while testing
	ctx := context.WithValue(context.Background(), "testing", true)
	for _, sub := range app.Commands {
		fmt.Println("RUN>", app.Name, sub.Name)
		if err := app.RunContext(ctx, []string{app.Name, sub.Name}); err != nil {
			t.Error(err)
		}
	}
}
