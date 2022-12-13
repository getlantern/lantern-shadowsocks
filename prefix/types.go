package prefix

import "io"

// MakePrefixFunc returns a function that returns a prefix
type MakePrefixFunc func() ([]byte, error)

// AbsorbPrefixFunc returns a function that absorbs a prefix from an io.Reader
// and return it
type AbsorbPrefixFunc func(io.Reader) ([]byte, error)
