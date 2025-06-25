package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

func Load() error {
	viper.AddConfigPath("${HOME}/.config/gsm/")
	viper.SetConfigName("server")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var fileNotFound viper.ConfigFileNotFoundError
		if errors.As(err, &fileNotFound) {
			slog.Warn("config file not found, ensure env vars are set")
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("reading config: %v", err)
		}
	}
	return nil
}
