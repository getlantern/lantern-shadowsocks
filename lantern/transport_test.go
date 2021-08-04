package lantern

import (
	"testing"

	"github.com/getlantern/transports/pluggable"
)

func TestTransport(t *testing.T) {
	pluggable.TestTransport(
		t, Transport{},
		&listenerConfig{
			CommonListenerConfig: pluggable.CommonListenerConfig{
				Addr: "localhost:0",
			},
			Secret: "test-secret",
			Cipher: "chacha20-ietf-poly1305",
		},
		func(listenerAddr string) interface{} {
			return &dialerConfig{
				CommonDialerConfig: pluggable.CommonDialerConfig{
					Addr: listenerAddr,
				},
				Secret:   "test-secret",
				Cipher:   "chacha20-ietf-poly1305",
				Upstream: "local",
			}
		},
	)
}
