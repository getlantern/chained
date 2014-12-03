package chained

import (
	"net"
	"testing"

	"github.com/getlantern/proxy"
)

func TestSuccessNotPipelined(t *testing.T) {
	doTest(t, false)
}

func TestSuccessPipelined(t *testing.T) {
	doTest(t, true)
}

func doTest(t *testing.T, pipelined bool) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Unable to listen: %s", err)
	}

	s := &Server{
		Dial: net.Dial,
	}
	go func() {
		err := s.Serve(l)
		if err != nil {
			t.Fatalf("Unable to serve: %s", err)
		}
	}()

	dialer := &Client{
		DialServer: func() (net.Conn, error) {
			return net.Dial(l.Addr().Network(), l.Addr().String())
		},
		Pipelined: false,
	}

	proxy.Test(t, dialer)
}
