package prefix

import (
	cryptoRand "crypto/rand"
	"fmt"
)

func MakeDNSOverTCPPrefix(length int) ([]byte, error) {
	if length >= 0xffff {
		return nil, fmt.Errorf("Invalid length %d", length)
	}

	b := make([]byte, 2)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate random bytes for DNS-over-TCP prefix: %w", err)
	}
	prefix := []byte{
		byte(length >> 8), byte(length), // Length
		b[0], b[1], // Transaction ID
		0x01, 0x20, // Flags: Standard query, recursion desired
	}
	return prefix, nil
}
