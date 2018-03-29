package commands

import "os"

const (
	defaultconfigA = "config.yml"
	defaultconfigB = "config.yaml"
)

// DefaultConfig - sets config based on ENV variable or default config.yml
func defaultConfig() string {
	if env := os.Getenv(configENV); env != "" {
		return env
	}

	if _, err := os.Stat(defaultconfigB); err == nil {
		return defaultconfigB
	}

	return defaultconfigA

}
