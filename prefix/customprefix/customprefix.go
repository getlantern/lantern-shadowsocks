// Package customprefix defines a prefix.Prefix type based on custom specifiers.
package customprefix

// CustomPrefix implements the prefix.Prefix interface using a custom specifier. See New for more
// details.
type CustomPrefix struct {
	// The generator program provided to New, parsed into a Go function.
	generator func() ([]byte, error)
}

// New creates a new Prefix based on a generator. The generator is a rudimentary program, with a
// grammar defined in EBNF as:
//
//	letter = "A" | "B" | "C" | "D" | "E" | "F" | "G"
//	  | "H" | "I" | "J" | "K" | "L" | "M" | "N"
//	  | "O" | "P" | "Q" | "R" | "S" | "T" | "U"
//	  | "V" | "W" | "X" | "Y" | "Z" | "a" | "b"
//	  | "c" | "d" | "e" | "f" | "g" | "h" | "i"
//	  | "j" | "k" | "l" | "m" | "n" | "o" | "p"
//	  | "q" | "r" | "s" | "t" | "u" | "v" | "w"
//	  | "x" | "y" | "z" ;
//	digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
//	whitespace = " " | "\t" | "\n" | "\r" ;
//	char = letter | digit | whitespace
//
//	number = digit, { digit } ;
//	string = char, { char } ;
//	identifier = letter, { letter | "_" } ;
//
//	version = "v", number , ".", number ;
//	arg list = string, { ",", " ", string } ;
//	function call = "$", identifier, "(", [ arg list ], ")" ;
//
//	generator = version, " ", { string | whitespace | function call }
//
// where the function calls reference a limited set of built-in functions defined in Builtins.
// Nested function calls are not supported.
//
// Generators begin with a version specifier. This specifier is used to make parsing decisions, but
// is left out of the output prefix.
//
// Example generators:
//
//	httpGet = `v1.0 GET /$random_string(5, 10) HTTP/1.1`
//
//	dnsOverTCP = `v1.0 $hex(05DC)$random_bytes(2, 3)$hex(0120)`
func New(generator string) (*CustomPrefix, error) {
	genFunc, err := parse(generator, Builtins)
	if err != nil {
		return nil, err
	}
	return &CustomPrefix{genFunc}, nil
}

func (p *CustomPrefix) Make() ([]byte, error) {
	return p.generator()
}
