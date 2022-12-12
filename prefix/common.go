package prefix

import (
	"bytes"
	"fmt"
	"io"
)

func AbsorbPrefixFromReader(clientReader io.Reader, expectedPrefix []byte) error {
	// Read the prefix
	actualPrefix := make([]byte, len(expectedPrefix))
	_, err := io.ReadFull(clientReader, actualPrefix)
	if err != nil {
		return err
	}
	// fmt.Printf("Read actualPrefix: %x\n", actualPrefix)

	// Assert the prefix
	if !bytes.Equal(actualPrefix, expectedPrefix) {
		return fmt.Errorf("Invalid prefix. Expected %x but got %x",
			expectedPrefix, actualPrefix)
	}
	return nil
}
