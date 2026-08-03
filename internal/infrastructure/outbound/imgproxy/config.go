package imgproxy

import (
	"errors"
	"strings"
)

type Config struct {
	PublicBaseURL string `koanf:"public-base-url"`
	Bucket        string `koanf:"bucket"`
	KeyHex        string `koanf:"key-hex"`
	SaltHex       string `koanf:"salt-hex"`
	Key           []byte
	Salt          []byte
}

func (c *Config) ApplyDefaults() {
	c.PublicBaseURL = strings.TrimRight(c.PublicBaseURL, "/")
}

func (c *Config) Validate() error {
	if c.PublicBaseURL == "" {
		return errors.New("imgproxy public base URL is required")
	}
	if len(c.Key) == 0 {
		return errors.New("imgproxy key is required")
	}
	if len(c.Salt) == 0 {
		return errors.New("imgproxy salt is required")
	}
	if c.Bucket == "" {
		return errors.New("imgproxy bucket is required")
	}
	return nil
}
