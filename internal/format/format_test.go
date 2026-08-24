package format

import (
	"encoding/json"
	"testing"

	"github.com/codebahn/codebahn-cli/internal/output"
)

func TestRegisterAndGet(t *testing.T) {
	called := false
	f := func(_ json.RawMessage, _ any, _ *output.Printer) error {
		called = true
		return nil
	}
	Register("test_tool", f)
	defer func() { delete(registry, "test_tool") }()

	got, ok := Get("test_tool")
	if !ok {
		t.Fatal("Get returned false for registered tool")
	}
	if err := got(nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("returned formatter was not the one registered")
	}
}

func TestGetMissing(t *testing.T) {
	_, ok := Get("nonexistent_tool")
	if ok {
		t.Error("Get returned true for unregistered tool")
	}
}
