package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

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

func startTCPEchoServer() (*net.TCPListener, *sync.WaitGroup, error) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		return nil, nil, fmt.Errorf("net.ListenTCP failed: %v", err)
	}
	var running sync.WaitGroup
	running.Add(1)
	go func() {
		defer running.Done()
		for {
			clientConn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Printf("ERROR: AcceptTCP failed: %v\n", err)
				return
			}
			running.Add(1)
			go func() {
				defer running.Done()
				io.Copy(clientConn, clientConn)
				clientConn.Close()
			}()
		}
	}()
	return listener, &running, nil
}

// connDialer is a net.Dialer that always returns the same connection.
type connDialer struct {
	c net.Conn
}

func (cd connDialer) Dial(network, addr string) (net.Conn, error) {
	return cd.c, nil
}
