package cryptocompare

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSymbols(t *testing.T) {
	k := viper.GetString("cryptocompare")
	p, err := GetSymbols(k, []string{"BTC", "ETH"}, []string{"USD", "EUR"})
	assert.NoError(t, err)
	assert.True(t, p.BTC.USD > 0)
	assert.True(t, p.ETH.USD > 0)
}
