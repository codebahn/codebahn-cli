package gen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/internal/context_detect"
	"github.com/codebahn/codebahn-cli/internal/format"
	"github.com/codebahn/codebahn-cli/internal/output"
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
func ExecuteAndPrint(cmd *cobra.Command, td tools.ToolDef, args any) error {
	c := ClientFrom(cmd.Context())
	if c == nil {
		return fmt.Errorf("not logged in; run 'codebahn auth login' first")
	}

	if err := resolveRepoContext(cmd, args); err != nil {
		return err
	}

	raw, err := c.Execute(cmd.Context(), td, args)
	if err != nil {
		return err
	}

	jsonMode, _ := cmd.Flags().GetBool("json")
	if jsonMode {
		if raw == nil {
			return nil
		}
		return printJSON(raw)
	}

	if f, ok := format.Get(td.Name); ok {
		return f(raw, args, output.NewPrinter(os.Stdout, false))
	}

	if raw == nil {
		return nil
	}
	return printJSON(raw)
}

func resolveRepoContext(cmd *cobra.Command, args any) error {
	v := reflect.ValueOf(args)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	ownerField := v.FieldByName("Owner")
	repoField := v.FieldByName("Repo")
	if !ownerField.IsValid() || !repoField.IsValid() {
		return nil
	}
	if !ownerField.CanSet() || !repoField.CanSet() {
		return nil
	}
	if ownerField.String() != "" && repoField.String() != "" {
		return nil
	}

	ownerSet := ownerField.String() != ""
	repoSet := repoField.String() != ""

	if ownerSet != repoSet {
		if ownerSet {
			return fmt.Errorf("--owner also requires --repo; use both or neither to auto-detect from git remote")
		}
		return fmt.Errorf("--repo also requires --owner; use both or neither to auto-detect from git remote")
	}

	instanceURL, _ := cmd.Flags().GetString("instance")
	ctx, ok := context_detect.DetectRepo(instanceURL)
	if !ok {
		return fmt.Errorf("could not detect repository from git remote; use --owner and --repo flags")
	}

	ownerField.SetString(ctx.Owner)
	repoField.SetString(ctx.Repo)
	return nil
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
