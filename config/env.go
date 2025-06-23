package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load() error {
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config: %v", err)
	}
	return nil
}
