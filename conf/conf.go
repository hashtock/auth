package conf

import (
	"net/url"
)

// Config struct holds information vital for correctly wroking service
// Providing all of them is mandatory
type Config struct {
	AppAddress         *url.URL
	ServeAddress       string
	SessionName        string
	SessionSecret      string
	GoogleClientID     string
	GoogleClientSecret string
}

var cfg *Config

// GetConfig returns global instance of configuration
func GetConfig() *Config {
	if cfg == nil {
		loadConfig()
	}

	return cfg
}
