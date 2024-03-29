// cmd/clienttest.go runs a client that connects to a running Shadowsocks server
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

	"github.com/Jigsaw-Code/outline-ss-server/client"
)

var (
	hostFlag   = flag.String("host", "", "Host to connect to")
	portFlag   = flag.Int("port", 0, "Port to connect to")
	secretFlag = flag.String("secret", "", "Secret to use")
	cipherFlag = flag.String("cipher", "chacha20-ietf-poly1305", "Cipher to use")
	prefixFlag = flag.String("prefix", "", "Prefix to use. Write as a hex string, e.g. AABBCC for []byte{0xAA, 0xBB, 0xCC}")
)

func main() {
	flag.Parse()

	var prefix []byte
	if *prefixFlag != "" {
		_, err := fmt.Sscanf(*prefixFlag, "%x", &prefix)
		if err != nil {
			panic(err)
		}
	}

	cl, err := client.NewClient(*hostFlag, *portFlag, *secretFlag, *cipherFlag)
	if err != nil {
		panic(err)
	}
	cl.SetTCPSaltGenerator(client.NewPrefixSaltGenerator(func() ([]byte, error) {
		return prefix, nil
	}))

	// Start a TCP connection against Google
	conn, err := cl.DialTCP(
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
