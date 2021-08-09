package goisbn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate10(t *testing.T) {
	type TestCase struct {
		name   string
		desc   string
		isbn   string
		expRes bool
	}
	testCases := []TestCase{
		{
			name:   "Happy Case",
			desc:   "valid isbn 10",
			isbn:   "0099588986",
			expRes: true,
		},
		{
			name:   "Sad Case",
			desc:   "invalid isbn 10, last digit X",
			isbn:   "009958898X",
			expRes: false,
		},
		{
			name:   "Sad Case",
			desc:   "invalid length",
			isbn:   "00995889866",
			expRes: false,
		},
		{
			name:   "Sad Case",
			desc:   "invalid string char",
			isbn:   "00995C8986",
			expRes: false,
		},
	}

	for _, v := range testCases {
		actRes := validate10(v.isbn)
		assert.Equal(t, v.expRes, actRes)
	}
}

func TestValidate13(t *testing.T) {
	type TestCase struct {
		name   string
		desc   string
		isbn   string
		expRes bool
	}
	testCases := []TestCase{
		{
			name:   "Happy Case",
			desc:   "valid isbn 13",
			isbn:   "9780099588986",
			expRes: true,
		},
		{
			name:   "Sad Case",
			desc:   "invalid length",
			isbn:   "97800995889866",
			expRes: false,
		},
		{
			name:   "Sad Case",
			desc:   "invalid string char",
			isbn:   "97800995C8986",
			expRes: false,
		},
	}

	for _, v := range testCases {
		actRes := validate13(v.isbn)
		assert.Equal(t, v.expRes, actRes)
	}
}
