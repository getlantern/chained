package chained

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"strings"
)

// Client creates a dialer function for a client proxy, using the given
// dialServer function to dial the server proxy. The dialer function issues a
// CONNECT request to instruct the server to connect to the specified network,
// addr.
//
// If pipelined is true, the dialer function will return before receiving a
// response to the CONNECT request. If pipelined is false, the dialer function
// will wait for and check the response to the CONNECT request before returning.
func Client(pipelined bool, dialServer func() (net.Conn, error)) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		conn, err := dialServer()
		if err != nil {
			return nil, fmt.Errorf("Unable to dial server at %s", err)
		}
		err = sendCONNECT(network, addr, conn, pipelined)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}
}

func sendCONNECT(network, addr string, conn net.Conn, pipelined bool) error {
	if !strings.Contains(network, "tcp") {
		return fmt.Errorf("%s connections are not supported, only tcp is supported", network)
	}

	req, err := http.NewRequest(CONNECT, addr, nil)
	if err != nil {
		return fmt.Errorf("Unable to construct CONNECT request: %s", err)
	}
	req.Host = addr
	err = req.Write(conn)
	if err != nil {
		return fmt.Errorf("Unable to write CONNECT request: %s", err)
	}

	r := bufio.NewReader(conn)
	if pipelined {
		go func() {
			err := checkCONNECTResponse(r, req)
			if err != nil {
				conn.Close()
				log.Error(err)
			}
		}()
	} else {
		err = checkCONNECTResponse(r, req)
	}
	return err
}

func checkCONNECTResponse(r *bufio.Reader, req *http.Request) error {
	resp, err := http.ReadResponse(r, req)
	if err != nil {
		return fmt.Errorf("Error reading CONNECT response: %s", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Bad status code on CONNECT response: %d", resp.StatusCode)
	}
	return nil
}
