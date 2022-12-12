// cmd/client.go runs a client that connects to a running Shadowsocks server
// and asserts that a TCP connection can be made to Google (as a stable
// example) and that the response is what we expect.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/getlantern/lantern-shadowsocks/client"
)

var (
	hostFlag   = flag.String("host", "", "Host to connect to")
	portFlag   = flag.Int("port", 0, "Port to connect to")
	secretFlag = flag.String("secret", "", "Secret to use")
	cipherFlag = flag.String("cipher", "chacha20-ietf-poly1305", "Cipher to use")
)

func main() {
	flag.Parse()

	client, err := client.NewClient(*hostFlag, *portFlag, *secretFlag, *cipherFlag)
	if err != nil {
		panic(err)
	}

	// Start a TCP connection against Google
	conn, err := client.DialTCP(
		nil,
		"142.250.181.206:80", // Google
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Send a request
	httpClient := http.Client{Transport: &http.Transport{Dial: connDialer{conn}.Dial}}
	resp, err := httpClient.Get("http://www.google.com/humans.txt")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response and assert that it's what we expect
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(string(buf), "Google") {
		panic("Unexpected response: " + string(buf))
	}

	fmt.Println("Success!")
}

// connDialer is a net.Dialer that always returns the same connection.
type connDialer struct {
	c net.Conn
}

func (cd connDialer) Dial(network, addr string) (net.Conn, error) {
	return cd.c, nil
}
