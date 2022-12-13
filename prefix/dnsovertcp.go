package prefix

import (
	cryptoRand "crypto/rand"
	"fmt"
	"io"
)

// See here for more info about this number:
// https://github.com/getlantern/lantern-internal/issues/4428#issuecomment-1337979698
//
// In short, it's a length that worked for Shadowsocks on Iran's MCI ISP. It's
// just easier to hardcode it until we need to make it variable
const dnsOverTCPMsgLen = 1500

func MakeDNSOverTCPPrefix() ([]byte, error) {
	if dnsOverTCPMsgLen >= 0xffff {
		return nil, fmt.Errorf("Invalid length %d", dnsOverTCPMsgLen)
	}

	b := make([]byte, 2)
	_, err := cryptoRand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate random bytes for DNS-over-TCP prefix: %w", err)
	}
	len := dnsOverTCPMsgLen
	prefix := []byte{
		byte(len >> 8), byte(len), // Length
		b[0], b[1], // Transaction ID
		0x01, 0x20, // Flags: Standard query, recursion desired
	}
	return prefix, nil
}

func AbsorbDNSOverTCPPrefix(clientReader io.Reader) ([]byte, error) {
	// Read the prefix
	actualPrefix := make([]byte, 6)
	_, err := io.ReadFull(clientReader, actualPrefix)
	if err != nil {
		return nil, fmt.Errorf("Unable to read DNS-over-TCP prefix: %w", err)
	}
	// fmt.Printf("Read actualPrefix: %x\n", actualPrefix)

	// Check the length
	actualLength := int(actualPrefix[0])<<8 | int(actualPrefix[1])
	if actualLength != dnsOverTCPMsgLen {
		return nil, fmt.Errorf("Invalid prefix length (expected %d; got %d)", dnsOverTCPMsgLen, actualLength)
	}

	// We're ignoring the transaction ID since that should be random

	// Check the flags
	if actualPrefix[4] != 0x01 || actualPrefix[5] != 0x20 {
		return nil, fmt.Errorf("Invalid prefix flags (expected 0x0120; got 0x%02x%02x)",
			actualPrefix[4], actualPrefix[5])
	}

	return actualPrefix, nil
}
