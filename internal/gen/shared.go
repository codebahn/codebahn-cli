package gen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/tools"
)

type contextKey int

const clientKey contextKey = iota

// WithClient stores a Client in the context for generated commands.
func WithClient(ctx context.Context, c *client.Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

// ClientFrom retrieves the Client from the context.
func ClientFrom(ctx context.Context) *client.Client {
	c, _ := ctx.Value(clientKey).(*client.Client)
	return c
}

// ExecuteAndPrint runs a tool's API call and prints the result.
// TTY output gets indented JSON; piped output gets compact JSON.
func ExecuteAndPrint(ctx context.Context, td tools.ToolDef, args any) error {
	c := ClientFrom(ctx)
	if c == nil {
		return fmt.Errorf("not logged in; run 'codebahn auth login' first")
	}

	raw, err := c.Execute(ctx, td, args)
	if err != nil {
		return err
	}

	if raw == nil {
		return nil
	}

	return printJSON(raw)
}

func printJSON(raw json.RawMessage) error {
	fi, _ := os.Stdout.Stat()
	isTTY := fi != nil && fi.Mode()&os.ModeCharDevice != 0

	if isTTY {
		var pretty json.RawMessage
		if err := json.Unmarshal(raw, &pretty); err == nil {
			indented, err := json.MarshalIndent(pretty, "", "  ")
			if err == nil {
				raw = indented
			}
		}
	}

	_, err := fmt.Fprintf(os.Stdout, "%s\n", raw)
	return err
}
