package chained

import (
	"net"
	"testing"

	"github.com/getlantern/proxytest"
)

func TestSuccessNotPipelined(t *testing.T) {
	doTest(t, false)
}

func TestSuccessPipelined(t *testing.T) {
	doTest(t, true)
}

func doTest(t *testing.T, pipelined bool) {
	// Set up listener for server proxy
	pl, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Unable to listen: %s", err)
	}

	// upstream proxy server
	s := &Server{
		Dial: net.Dial,
	}
	go func() {
		err := s.Serve(pl)
		if err != nil {
			t.Fatalf("Unable to serve: %s", err)
		}
	}()

	// proxy client
	dial := Client(false, func() (net.Conn, error) {
		return net.Dial(pl.Addr().Network(), pl.Addr().String())
	})

	proxytest.Go(t, dial)
}
