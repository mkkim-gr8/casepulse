package healthcheck

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
)

func TestLocalAddr(t *testing.T) {
	tests := map[string]string{
		":9090":          "localhost:9090",
		"0.0.0.0:9090":   "localhost:9090",
		"[::]:9090":      "localhost:9090",
		"127.0.0.1:9091": "127.0.0.1:9091",
		"clio:9091":      "clio:9091",
	}
	for input, want := range tests {
		if got := LocalAddr(input); got != want {
			t.Errorf("LocalAddr(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestProbeGRPC(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	RegisterGRPC(server)
	go server.Serve(listener)
	t.Cleanup(server.Stop)

	if err := ProbeGRPC(listener.Addr().String()); err != nil {
		t.Fatalf("ProbeGRPC on running server: %v", err)
	}

	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := closed.Addr().String()
	closed.Close()
	if err := ProbeGRPC(addr); err == nil {
		t.Fatal("ProbeGRPC on closed port succeeded, want error")
	}
}

func TestProbeHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	addr := server.Listener.Addr().String()

	if err := ProbeHTTP(addr, "/healthz"); err != nil {
		t.Fatalf("ProbeHTTP(/healthz): %v", err)
	}
	if err := ProbeHTTP(addr, "/missing"); err == nil {
		t.Fatal("ProbeHTTP(/missing) succeeded, want error")
	}
}

func TestIsHealthMethod(t *testing.T) {
	if !IsHealthMethod("/grpc.health.v1.Health/Check") {
		t.Error("Check should be a health method")
	}
	if IsHealthMethod("/casepulse.clio.v1.ClioService/ListCases") {
		t.Error("ListCases should not be a health method")
	}
}
