package event_test

import (
	"github.com/sgerogia/sol-stablecoin/tpp-client/event"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestToDecimal(t *testing.T) {
	cases := []struct {
		v   string
		exp string
	}{
		{v: "1000000000000000000", exp: "1"},
		{v: "123000000000000000000", exp: "123"},
		{v: "1200000000000000000", exp: "1.2"},
		{v: "12345000000000000000000", exp: "12345"},
		{v: "10000000000000000", exp: "0.01"},
	}
	for _, c := range cases {
		x, _ := new(big.Int).SetString(c.v, 10)
		got := event.ToDecimal(x, event.DECIMAL_DIGITS)
		assert.Equal(t, c.exp, got)
	}
}

func TestToWei(t *testing.T) {
	cases := []struct {
		v   string
		exp string
	}{
		{v: "1", exp: "1000000000000000000"},
		{v: "123", exp: "123000000000000000000"},
		{v: "1.2", exp: "1200000000000000000"},
		{v: "12345", exp: "12345000000000000000000"},
		{v: "0.01", exp: "10000000000000000"},
	}
	for _, c := range cases {
		x, _ := new(big.Int).SetString(c.exp, 10)
		got := event.ToWei(c.v, event.DECIMAL_DIGITS)
		assert.Equal(t, x, got)
	}
}
