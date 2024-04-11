package config

import (
	"github.com/pelletier/go-toml/v2"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	Title      string
	BankClient struct {
		ClientId     string
		ClientSecret string
		RedirectUrl  string
	}
	Ethereum struct {
		ProviderUrl     string
		ChainId         int64
		ContractAddress string
		MaxGas          int64
	}
	Tuning struct {
		CronSchedule      int
		BankClientTimeout int
	}
	BankAccount struct {
		SortCode      string
		AccountNumber string
		AccountName   string
	}
}

// LoadConfigData loads the configuration from the given TOML data
func LoadConfigData(data []byte, l *zap.SugaredLogger) (*Config, error) {
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadConfig loads the configuration from the given TOML file
func LoadConfig(path string, l *zap.SugaredLogger) (*Config, error) {
	if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		return LoadConfigData(data, l)
	}
}
