package lantern

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	mrand "math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/getlantern/transports/pluggable"

	ssclient "github.com/getlantern/lantern-shadowsocks/client"
)

type dialerConfig struct {
	pluggable.CommonDialerConfig
	Secret, Cipher, Upstream string
}

type listenerConfig struct {
	pluggable.CommonListenerConfig
	Secret, Cipher string
	ReplayHistory  int
}

// Transport implements getlantern/transports/pluggable.Transport.
type Transport struct{}

func (_ Transport) NewDialer(config interface{}, _ string, _ pluggable.ClientConfig) (pluggable.Dialer, error) {
	cfg, ok := config.(*dialerConfig)
	if !ok {
		return nil, fmt.Errorf("expected config of type %T, but got %T", &dialerConfig{}, config)
	}

	gen, err := newShadowsocksUpstreamGenerator(cfg.Upstream)
	if err != nil {
		return nil, fmt.Errorf("failed to init upstream generator: %w", err)
	}

	host, _port, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("malformed address; failed to parse: %w", err)
	}
	port, err := strconv.Atoi(_port)
	if err != nil {
		return nil, fmt.Errorf("malformed address; failed to parse port as int: %w", err)
	}
	client, err := ssclient.NewClient(host, port, cfg.Secret, cfg.Cipher)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shadowsocks client: %w", err)
	}

	return pluggable.WrapDialer(func() (net.Conn, error) {
		conn, err := client.DialTCP(nil, gen.newUpstream())
		if err != nil {
			return nil, err
		}
		return &ssWrapConn{conn}, nil
	}), nil
}

func (_ Transport) NewListener(config interface{}, _ pluggable.RuntimeListenerConfig) (net.Listener, error) {
	cfg, ok := config.(*listenerConfig)
	if !ok {
		return nil, fmt.Errorf("expected config of type %T, but got %T", &listenerConfig{}, config)
	}

	configs := []CipherConfig{
		{
			ID:     "default",
			Secret: cfg.Secret,
			Cipher: cfg.Cipher,
		},
	}
	ciphers, err := NewCipherListWithConfigs(configs)
	if err != nil {
		return nil, fmt.Errorf("Unable to create shadowsocks cipher: %w", err)
	}
	l, err := ListenLocalTCP(cfg.Addr, ciphers, cfg.ReplayHistory)
	if err != nil {
		return nil, fmt.Errorf("Unable to listen for shadowsocks: %w", err)
	}
	return l, nil
}

func (_ Transport) DialerConfig() interface{}   { return &dialerConfig{} }
func (_ Transport) ListenerConfig() interface{} { return &listenerConfig{} }

type shadowsocksUpstreamGenerator struct {
	baseUpstream string
	rng          *rand.Rand
	rngmx        sync.Mutex
}

func newShadowsocksUpstreamGenerator(baseUpstream string) (*shadowsocksUpstreamGenerator, error) {
	var seed int64
	err := binary.Read(crand.Reader, binary.BigEndian, &seed)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize rng: %w", err)
	}
	source := mrand.NewSource(seed)
	rng := mrand.New(source)
	return &shadowsocksUpstreamGenerator{baseUpstream, rng, sync.Mutex{}}, nil
}

// newUpstream() creates a marker upstream address.  This isn't an
// acutal upstream that will be dialed, it signals that the upstream
// should be determined by other methods.  It's just a bit random just to
// mix it up and not do anything especially consistent on every dial.
//
// To satisy shadowsocks expectations, a small random string is prefixed onto the
// configured suffix (along with a .) and a port is affixed to the end.
func (gen *shadowsocksUpstreamGenerator) newUpstream() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	gen.rngmx.Lock()
	defer gen.rngmx.Unlock()
	// [2 - 22]
	sz := 2 + gen.rng.Intn(21)
	b := make([]byte, sz)
	for i := range b {
		b[i] = letters[gen.rng.Intn(len(letters))]
	}
	return fmt.Sprintf("%s.%s:443", string(b), gen.baseUpstream)
}

// this is a helper to smooth out error bumps
// that the rest of lantern doesn't really expect, but happen
// in the shadowsocks impl when closing.
type ssWrapConn struct {
	net.Conn
}

func (c *ssWrapConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if errors.Is(err, net.ErrClosed) {
		err = io.EOF
	}
	return n, err
}

func (c *ssWrapConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if errors.Is(err, net.ErrClosed) {
		err = io.EOF
	}
	return n, err
}
