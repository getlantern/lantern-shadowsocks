package customprefix

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorRegexp(t *testing.T) {
	for _, testCase := range []struct {
		input, majorVersion, minorVersion, rest string
		shouldMatch                             bool
	}{
		{
			`v1.0 HTTP/1.1`, `1`, `0`, `HTTP/1.1`, true,
		},
		{
			`v100.100 HTTP/1.1`, `100`, `100`, `HTTP/1.1`, true,
		},
		{
			"v1.0 MyProtocol/v1.0 blah blah $foo(3, 4)\n blah$bar()",
			"1", "0", "MyProtocol/v1.0 blah blah $foo(3, 4)\n blah$bar()", true,
		},
		{
			`v10 HTTP/1.1`, "", "", "", false,
		},
		{
			`v1.0HTTP/1.1`, "", "", "", false,
		},
		{
			`v1.0`, "", "", "", false,
		},
	} {
		t.Run(testCase.input, func(t *testing.T) {
			doesMatch := generatorRegexp.MatchString(testCase.input)
			if !assert.Equal(t, testCase.shouldMatch, doesMatch) || !doesMatch {
				return
			}

			submatches := generatorRegexp.FindStringSubmatch(testCase.input)
			if !assert.Len(t, submatches, 4) {
				return
			}

			assert.Equal(t, testCase.majorVersion, submatches[1])
			assert.Equal(t, testCase.minorVersion, submatches[2])
			assert.Equal(t, testCase.rest, submatches[3])
		})
	}
}

func TestFunctionCallRegexp(t *testing.T) {
	t.Run("top level", func(t *testing.T) {
		for _, testCase := range []struct {
			input           string
			expectedMatches []string
		}{
			{
				`HTTP/1.1`, nil,
			},
			{
				"MyProtocol/v1.0 blah blah $foo(3, 4)\n blah$bar()",
				[]string{`$foo(3, 4)`, `$bar()`},
			},
		} {
			t.Run(testCase.input, func(t *testing.T) {
				matches := functionCallRegexp.FindAllString(testCase.input, -1)
				assert.Equal(t, testCase.expectedMatches, matches)
			})
		}
	})

	t.Run("sub match", func(t *testing.T) {
		for _, testCase := range []struct {
			input, functionName, args string
		}{
			{`$foo(3, 4)`, "foo", "3, 4"},
			{`$bar()`, "bar", ""},
		} {
			t.Run(testCase.input, func(t *testing.T) {
				submatches := functionCallRegexp.FindStringSubmatch(testCase.input)
				if !assert.Len(t, submatches, 3) {
					return
				}

				assert.Equal(t, testCase.functionName, submatches[1])
				assert.Equal(t, testCase.args, submatches[2])
			})
		}
	})
}

func TestParse(t *testing.T) {
	// A special set of built-ins, designed to test arg parsing, result printing, and errors.
	builtins := map[string]builtin{
		"printB": func(args []string) ([]byte, error) {
			return []byte{'B'}, nil
		},
		"printC": func(_ []string) ([]byte, error) {
			return []byte{'C'}, nil
		},
		"printBinaryFF": func(_ []string) ([]byte, error) {
			return []byte{0xff}, nil
		},
		"assertNoArgs": func(args []string) ([]byte, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("expected no args, received %d", len(args))
			}
			return []byte{}, nil
		},
		"assertNine": func(args []string) ([]byte, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("expected a single argument, received %d", len(args))
			}
			if args[0] != "9" {
				return nil, fmt.Errorf("expected '9' as arg 0, received '%s'", args[0])
			}
			return []byte{}, nil
		},
		"assertNineAndEleven": func(args []string) ([]byte, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("expected a single argument, received %d", len(args))
			}
			if args[0] != "9" {
				return nil, fmt.Errorf("expected '9' as arg 0, received '%s'", args[0])
			}
			if args[1] != "11" {
				return nil, fmt.Errorf("expected '11' as arg 1, received '%s'", args[1])
			}
			return []byte{}, nil
		},
		"assertNineNineFACE": func(args []string) ([]byte, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("expected a single argument, received %d", len(args))
			}
			if args[0] != "09FACE" {
				return nil, fmt.Errorf("expected '09FACE' as arg 0, received '%s'", args[0])
			}
			return []byte{}, nil
		},
		"returnError": func(_ []string) ([]byte, error) {
			return nil, errors.New("runtime error")
		},
	}

	for _, testCase := range []struct {
		input                                string
		expectedOutput                       []byte
		expectParseError, expectRuntimeError bool
	}{
		{
			`v1.0 $printB()`,
			[]byte("B"), false, false,
		},
		{
			`v1.0 $printB()$printC()`,
			[]byte("BC"), false, false,
		},
		{
			`v1.0 A$printB()$printC()`,
			[]byte("ABC"), false, false,
		},
		{
			// n.b. printB ignores its arguments.
			`v1.0 A$printB(arg1, arg2, arg3)$printC()`,
			[]byte("ABC"), false, false,
		},
		{
			`v1.0 $printB()$printBinaryFF()`,
			[]byte{'B', 0xff}, false, false,
		},
		{
			"v1.0 \n$printB()\n$printBinaryFF()\r\n",
			[]byte{'\n', 'B', '\n', 0xff, '\r', '\n'}, false, false,
		},
		{
			"v1.0 \n$printB()\nSOME TEXT\nA $printB() $printC() D\r\n",
			[]byte("\nB\nSOME TEXT\nA B C D\r\n"), false, false,
		},
		{
			`v1.0 $assertNoArgs()`,
			[]byte{}, false, false,
		},
		{
			`v1.0 $assertNine(9)`,
			[]byte{}, false, false,
		},
		{
			`v1.0 $assertNineAndEleven(9, 11)`,
			[]byte{}, false, false,
		},
		{
			`v1.0 $assertNineNineFACE(09FACE)`,
			[]byte{}, false, false,
		},
		{
			`v1.0 $assertNoArgs()$assertNineAndEleven(9, 11)$assertNineNineFACE(09FACE)`,
			[]byte{}, false, false,
		},
		{
			// Version exceeds supported version.
			`v999999.0 $printB()$printC()`,
			nil, true, false,
		},
		{
			// No version specifier.
			`$printB()$printC()`,
			nil, true, false,
		},
		{
			// No space after version specifier
			`v1.0$printB()$printC()`,
			nil, true, false,
		},
		{
			// Unsupported built-in.
			`v1.0 $printB()$unsupportedBuiltin()`,
			nil, true, false,
		},
		{
			// Nested functions are unsupported.
			// n.b. printB ignores its arguments.
			`v1.0 $printB($printC())`,
			nil, true, false,
		},
		{
			// Runtime error returned by built-in.
			`v1.0 $printB()$printC()$returnError()`,
			nil, false, true,
		},
	} {
		// t.Run(strconv.Itoa(i), func(t *testing.T) {
		t.Run(testCase.input, func(t *testing.T) {
			genFunc, err := parse(testCase.input, builtins)
			if testCase.expectParseError {
				assert.Error(t, err)
				return
			}
			if !assert.NoError(t, err) {
				return
			}

			output, err := genFunc()
			if testCase.expectRuntimeError {
				assert.Error(t, err)
				return
			}
			if !assert.NoError(t, err) {
				return
			}

			// Guard against bugs which produce runaway output.
			if len(output) > 1024 {
				t.Logf("output length %d exceeds guard; skipping equality check", len(output))
				t.Fail()
				return
			}

			assert.Equal(t, string(testCase.expectedOutput), string(output))
		})
	}
}
