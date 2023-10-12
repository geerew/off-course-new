package security

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RandomString(t *testing.T) {
	testRandomString(t, RandomString)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RandomStringWithAlphabet(t *testing.T) {
	testRandomStringWithAlphabet(t, RandomStringWithAlphabet)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_PseudorandomString(t *testing.T) {
	testRandomString(t, PseudorandomString)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_PseudorandomStringWithAlphabet(t *testing.T) {
	testRandomStringWithAlphabet(t, PseudorandomStringWithAlphabet)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func testRandomStringWithAlphabet(t *testing.T, randomFunc func(n int, alphabet string) string) {
	tests := []struct {
		alphabet      string
		expectPattern string
	}{
		{"0123456789_", `[0-9_]+`},
		{"abcdef123", `[abcdef123]+`},
		{"!@#$%^&*()", `[\!\@\#\$\%\^\&\*\(\)]+`},
	}

	for _, tt := range tests {
		generated := make([]string, 0, 1000)
		length := 10

		for j := 0; j < 1000; j++ {
			res := randomFunc(length, tt.alphabet)
			require.Len(t, res, length)

			reg := regexp.MustCompile(tt.expectPattern)
			assert.True(t, reg.MatchString(res))

			for _, str := range generated {
				assert.NotEqual(t, res, str)
			}

			generated = append(generated, res)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func testRandomString(t *testing.T, randomFunc func(n int) string) {
	generated := make([]string, 0, 1000)
	reg := regexp.MustCompile(`[a-zA-Z0-9]+`)
	length := 10

	for i := 0; i < 1000; i++ {
		res := randomFunc(length)
		require.Len(t, res, length)
		assert.True(t, reg.MatchString(res))

		for _, str := range generated {
			assert.NotEqual(t, res, str)

		}

		generated = append(generated, res)
	}
}
