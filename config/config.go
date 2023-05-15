package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

// Conf holds the configuration for the application
type Conf struct {
	Server serverConf
}

// serverConf holds the server configuration
type serverConf struct {
	Addr         string        `env:"SERVER_ADDR,required"`
	Port         int           `env:"SERVER_PORT,required"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
}

// LoadAppConfig loads the application configuration from the environment
func LoadAppConfig() error {
	// viper read config from from autodealer/.env
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Errorf("Failed to read config file: %s", err)
		return err
	}

	requiredEnvVars := []string{"SERVER_ADDR", "SERVER_PORT", "SERVER_TIMEOUT_READ", "SERVER_TIMEOUT_WRITE", "SERVER_TIMEOUT_IDLE"}

	for _, envVar := range requiredEnvVars {
		if !viper.IsSet(envVar) {
			err := errors.New("Environment variable " + envVar + " is not set")
			logrus.Error(err)
			return err
		}
	}

	return nil
}
