package security

import (
	cryptoRand "crypto/rand"
	"math/big"
	mathRand "math/rand"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const defaultRandomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RandomString generates a cryptographically random string with the specified length. The
// generated string matches [A-Za-z0-9]+
//
// Slower than PseudorandomString, but cryptographically secure
func RandomString(length int) string {
	return RandomStringWithAlphabet(length, defaultRandomAlphabet)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RandomStringWithAlphabet generates a cryptographically random string with the specified length
// and characters set
//
// # It panics if for some reason rand.Int returns a non-nil error
//
// Slower than PseudorandomString, but cryptographically secure
func RandomStringWithAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range b {
		n, err := cryptoRand.Int(cryptoRand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = alphabet[n.Int64()]
	}

	return string(b)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PseudorandomString generates a pseudorandom string with the specified length. The generated
// string matches [A-Za-z0-9]+
//
// Faster than RandomString, but not cryptographically secure
func PseudorandomString(length int) string {
	return PseudorandomStringWithAlphabet(length, defaultRandomAlphabet)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PseudorandomStringWithAlphabet generates a pseudorandom string with the specified length
//
// Faster than RandomString, but not cryptographically secure
func PseudorandomStringWithAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	max := len(alphabet)

	for i := range b {
		b[i] = alphabet[mathRand.Intn(max)]
	}

	return string(b)
}
