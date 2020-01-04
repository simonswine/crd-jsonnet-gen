package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLowerName(t *testing.T) {
	for _, tc := range []struct {
		input string
		exp   string
	}{
		{
			"CAIssuer",
			"caIssuer",
		},
		{
			"CAIssuerID",
			"caIssuerID",
		},
		{
			"CA",
			"ca",
		},
		{
			"ACMEIssuer",
			"acmeIssuer",
		},
		{
			"OrderStatus",
			"orderStatus",
		},
	} {
		assert.Equal(
			t,
			tc.exp,
			lowerName(tc.input),
			fmt.Sprintf("lowerName(\"%s\")", tc.input),
		)
	}
}
