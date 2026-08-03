package token

import (
	"fmt"
)

type Config struct {
	JWTSecret string `koanf:"jwt-secret"`
}

func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("upload-token.jwt-secret is required")
	}
	return nil
}

func (c *Config) ApplyDefaults() {
}
