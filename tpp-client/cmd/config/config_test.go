package config_test

import (
	"github.com/sgerogia/sol-stablecoin/tpp-client/cmd/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	f "path/filepath"
	"runtime"
	"testing"
)

const CONF_FILE = "../../../tpp-client.toml"

func TestLoadConfigData(t *testing.T) {
	// find the example config file to load
	_, curr, _, _ := runtime.Caller(0)
	file := f.Join(f.Dir(curr), CONF_FILE)

	l := zaptest.NewLogger(t).Sugar()
	c, err := config.LoadConfig(file, l)
	require.NoError(t, err)

	assert.Equal(t, "Example TPP client configuration", c.Title)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", c.Ethereum.ContractAddress)
	assert.Equal(t, int64(300000), c.Ethereum.MaxGas)
	assert.Equal(t, int64(11155111), c.Ethereum.ChainId)
	assert.Equal(t, "http://localhost:8080/callback", c.BankClient.RedirectUrl)
	assert.Equal(t, 30, c.Tuning.BankClientTimeout)
	assert.Equal(t, 5, c.Tuning.BankCronSchedule)
	assert.Equal(t, 1, c.Tuning.ChainCronSchedule)
	assert.Equal(t, uint64(10), c.Tuning.StartingBlock)
	assert.Equal(t, "ProvableGBP Limited", c.BankAccount.AccountName)
}
