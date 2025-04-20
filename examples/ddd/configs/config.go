package configs

import (
	"github.com/spf13/viper"
)

func LoadConfig(file string) (*Config, error) {

	var cfg Config

	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
