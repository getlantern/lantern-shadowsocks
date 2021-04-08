package lantern

import (
	"bytes"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	logging "github.com/op/go-logging"

	"github.com/getlantern/lantern-shadowsocks/client"
	"github.com/getlantern/lantern-shadowsocks/service"
	"github.com/getlantern/lantern-shadowsocks/service/metrics"
	ss "github.com/getlantern/lantern-shadowsocks/shadowsocks"
	"github.com/stretchr/testify/require"
)

func init() {
	logging.SetLevel(logging.INFO, "")
}

func makeTestCiphers(secrets []string) (service.CipherList, error) {
	configs := make([]CipherConfig, len(secrets))
	for i, secret := range secrets {
		configs[i].Secret = secret
	}

	cipherList, err := NewCipherListWithConfigs(configs)
	return cipherList, err
}

// tests interception of upstream connection
func TestLocalUpstreamHandling(t *testing.T) {
	req := make([]byte, 1024)
	res := make([]byte, 2048)

	_, err := rand.Read(req)
	require.Nil(t, err, "Failed to generate random request")
	_, err = rand.Read(res)
	require.Nil(t, err, "Failed to generate random response")

	l0, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	require.Nil(t, err, "ListenTCP failed: %v", err)
	secrets := ss.MakeTestSecrets(1)
	cipherList, err := makeTestCiphers(secrets)
	require.Nil(t, err, "MakeTestCiphers failed: %v", err)
	testMetrics := &metrics.NoOpMetrics{}

	options := &ListenerOptions{
		Listener: l0,
		Ciphers:  cipherList,
		Metrics:  testMetrics,
		Timeout:  200 * time.Millisecond,
	}

	l1 := ListenLocalTCPOptions(options)
	defer l1.Close()

	go func() {
		for {
			c, err := l1.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 2*len(req))
				n, err := c.Read(buf)
				if err != nil {
					logger.Errorf("error reading: %v", err)
					return
				}
				buf = buf[:n]
				if !bytes.Equal(buf, req) {
					logger.Errorf("unexpected request %v %v", buf, req)
					return
				}
				c.Write(res)
			}(c)
		}
	}()

	host, portStr, _ := net.SplitHostPort(l1.Addr().String())
	port, err := strconv.ParseInt(portStr, 10, 32)
	require.Nil(t, err, "Error parsing port")
	client, err := client.NewClient(host, int(port), secrets[0], ss.TestCipher)
	require.Nil(t, err, "Error creating client")
	conn, err := client.DialTCP(nil, "127.0.0.1:443")
	require.Nil(t, err, "failed to dial")
	_, err = conn.Write(req)
	require.Nil(t, err, "failed to write request")

	buf := make([]byte, 2*len(res))
	n, err := conn.Read(buf)
	require.Nil(t, err, "failed to read response")
	require.Equal(t, res, buf[:n], "unexpected response")
}
