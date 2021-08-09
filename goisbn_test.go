package goisbn

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGoISBN(t *testing.T) {
	type testCase struct {
		name            string
		desc            string
		providers       []string
		lenProviders    int
		goodreadsAPIKey string
		isbnDBAPIKey    string
		lenResolvers    []string
	}
	testCases := []testCase{
		{
			name:            "Happy Case",
			desc:            "all ok",
			providers:       DEFAULT_PROVIDERS,
			lenProviders:    4,
			goodreadsAPIKey: "dummy goodread key",
			isbnDBAPIKey:    "dummy isbnapi key",
			lenResolvers:    []string{"goodreads", "google", "isbndb", "openlibrary"},
		},
		{
			name:            "Happy Case",
			desc:            "no providers supplied, defaulting to default providers",
			providers:       []string{},
			lenProviders:    4,
			goodreadsAPIKey: "dummy goodread key",
			isbnDBAPIKey:    "dummy isbnapi key",
			lenResolvers:    []string{"goodreads", "google", "isbndb", "openlibrary"},
		},
		{
			name:            "Happy Case",
			desc:            "goodreads api key not supplied, removing goodreads from provider list",
			providers:       DEFAULT_PROVIDERS,
			lenProviders:    3,
			goodreadsAPIKey: "",
			isbnDBAPIKey:    "dummy isbnapi key",
			lenResolvers:    []string{"google", "isbndb", "openlibrary"},
		},
		{
			name:            "Happy Case",
			desc:            "isbndb api key not supplied, removing isbndb from provider list",
			providers:       DEFAULT_PROVIDERS,
			lenProviders:    3,
			goodreadsAPIKey: "dummy goodread key",
			isbnDBAPIKey:    "",
			lenResolvers:    []string{"goodreads", "google", "openlibrary"},
		},
	}
	for _, v := range testCases {
		defer unsetEnv()
		os.Setenv(goodreadsAPIKey, v.goodreadsAPIKey)
		os.Setenv(isbndbAPIKey, v.isbnDBAPIKey)
		gi := NewGoISBN(v.providers)
		assert.Equal(t, len(gi.providers), v.lenProviders)
		assert.Equal(t, gi.goodreadAPIKey, v.goodreadsAPIKey)
		assert.Equal(t, gi.isbndbAPIKey, v.isbnDBAPIKey)
		for _, val := range v.lenResolvers {
			_, ok := gi.resolvers[val]
			assert.Equal(t, ok, true)
		}
	}
}

func TestValidateISBN(t *testing.T) {
	type testCase struct {
		name   string
		desc   string
		isbn   string
		expRes bool
	}
	testCases := []testCase{
		{
			name:   "Happy Case",
			desc:   "all ok, valid isbn 13",
			isbn:   "9780099588986",
			expRes: true,
		},
		{
			name:   "Happy Case",
			desc:   "all ok, valid isbn 10",
			isbn:   "0099588986",
			expRes: true,
		},
		{
			name:   "Sad Case",
			desc:   "invalid length",
			isbn:   "00995889862",
			expRes: false,
		},
	}

	gi := NewGoISBN(DEFAULT_PROVIDERS)

	for _, v := range testCases {

		actRes := gi.ValidateISBN(v.isbn)

		assert.Equal(t, v.expRes, actRes)
	}

}

func unsetEnv() (restore func()) {
	before := map[string]string{
		goodreadsAPIKey: os.Getenv(goodreadsAPIKey),
		isbndbAPIKey:    os.Getenv(isbndbAPIKey),
	}
	for k := range before {
		os.Unsetenv(k)
	}
	return func() {
		for k, v := range before {
			if os.Getenv(k) != v {
				os.Setenv(k, v)
			}
		}
	}
}
