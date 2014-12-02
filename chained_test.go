package chained

import (
	"io"
	"net"
	"testing"

	"github.com/getlantern/testify/assert"
)

var (
	ping = []byte("ping")
	pong = []byte("pong")
)

func TestSuccessNotPipelined(t *testing.T) {
	doTest(t, false)
}

func TestSuccessPipelined(t *testing.T) {
	doTest(t, true)
}

func doTest(t *testing.T, pipelined bool) {
	// Set up listener for server endpoint
	sl, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Unable to listen: %s", err)
	}

	// Set up listener for server proxy
	pl, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Unable to listen: %s", err)
	}

	// Server that responds to ping
	go func() {
		conn, err := sl.Accept()
		if err != nil {
			t.Fatalf("Unable to accept connection: %s", err)
			return
		}
		defer conn.Close()
		b := make([]byte, 4)
		_, err = io.ReadFull(conn, b)
		if err != nil {
			t.Fatalf("Unable to read from client: %s", err)
		}
		assert.Equal(t, ping, b, "Didn't receive correct ping message")
		_, err = conn.Write(pong)
		if err != nil {
			t.Fatalf("Unable to write to client: %s", err)
		}
	}()

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

	conn, err := dial(sl.Addr().Network(), sl.Addr().String())
	if err != nil {
		t.Fatalf("Unable to dial via proxy: %s", err)
	}
	defer conn.Close()

	_, err = conn.Write(ping)
	if err != nil {
		t.Fatalf("Unable to write to server via proxy: %s", err)
	}

	b := make([]byte, 4)
	_, err = io.ReadFull(conn, b)
	if err != nil {
		t.Fatalf("Unable to read from server: %s", err)
	}
	assert.Equal(t, pong, b, "Didn't receive correct pong message")
}
