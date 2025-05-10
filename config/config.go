package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	ConnectingString string `mapstructure:"CONNECTING_STRING"`
	Port             string `mapstructure:"PORT"`
	NetworkType      string `mapstructure:"NETWORK_TYPE"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("conf")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		_ = fmt.Errorf("do not parse config file:%v", err)
	}

	err = viper.Unmarshal(&config)

	return
}
