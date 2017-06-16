package commands

import "os"

// DefaultConfig - sets config based on ENV variable or default config.yml
func defaultConfig() string {
	env := os.Getenv(configENV)
	if env == "" {
		return "config.yml"
	}
	return env
}
