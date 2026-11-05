package oauth

import (
	"context"
	"testing"
	"time"
)

func TestStartCallbackServer_SetsReadHeaderTimeout(t *testing.T) {
	srv, _, _, err := startCallbackServer()
	if err != nil {
		t.Fatalf("startCallbackServer: %v", err)
	}
	defer func() { _ = srv.Shutdown(context.Background()) }()

	if srv.ReadHeaderTimeout <= 0 {
		t.Errorf("ReadHeaderTimeout = %v, want a positive timeout against slow-header clients", srv.ReadHeaderTimeout)
	}
	if srv.ReadHeaderTimeout > time.Minute {
		t.Errorf("ReadHeaderTimeout = %v, want at most a minute", srv.ReadHeaderTimeout)
	}
}
