package serviceutils

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestServe_StopsCleanlyWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- serve(ctx, grpc.NewServer(), "127.0.0.1:0",
			&http.Server{Addr: "127.0.0.1:0", ReadHeaderTimeout: time.Second})
	}()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("clean shutdown returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after cancellation")
	}
}

// Regression: a server failing at runtime used to log.Fatal inside its
// goroutine, skipping every deferred cleanup. It must be returned instead.
func TestServe_ReturnsServerFailure(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Close() }()

	// HTTP server on an address that is already taken fails immediately
	err = serve(context.Background(), nil, "",
		&http.Server{Addr: busy.Addr().String(), ReadHeaderTimeout: time.Second})
	if err == nil {
		t.Fatal("expected an error for a port that is already in use")
	}
}

func TestServe_ReturnsListenError(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Close() }()

	if err := serve(context.Background(), grpc.NewServer(), busy.Addr().String(), nil); err == nil {
		t.Fatal("expected a listen error for a port that is already in use")
	}
}
