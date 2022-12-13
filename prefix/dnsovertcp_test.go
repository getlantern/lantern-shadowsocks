package prefix

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDNSOverTCPPrefix(t *testing.T) {
	for i := 0; i < 100; i++ {
		prefix, err := MakeDNSOverTCPPrefix()
		require.NoError(t, err)
		b, err := AbsorbDNSOverTCPPrefix(bytes.NewReader(prefix))
		require.NoError(t, err)
		require.Equal(t, prefix, b)
	}
}
