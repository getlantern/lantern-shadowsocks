package prefix

import (
	"fmt"
	"io"
)

func AbsorbPrefixFromReader(
	clientReader io.Reader,
	expectedPrefixSize int) (absorbedPrefixBuf []byte, err error) {
	// Read the prefix
	absorbedPrefixBuf = make([]byte, expectedPrefixSize)
	if _, err = io.ReadFull(clientReader, absorbedPrefixBuf); err != nil {
		return nil, fmt.Errorf("Failed to read prefix: %s", err)
	}
	// fmt.Printf("Read actualPrefix: %x\n", actualPrefix)

	// Assert the prefix
	// if !bytes.Equal(actualPrefix, expectedPrefix) {
	// 	return fmt.Errorf("Invalid prefix. Expected %x but got %x",
	// 		expectedPrefix, actualPrefix)
	// }
	return absorbedPrefixBuf, nil
}
