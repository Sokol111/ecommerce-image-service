package imgproxy

import (
	"encoding/hex"
	"errors"
	"strings"
)

type Config struct {
	PublicBaseURL  string `koanf:"public-base-url"` // IMGPROXY_PUBLIC_BASE_URL
	Bucket         string `koanf:"bucket"`          // S3_BUCKET
	KeyHex         string `koanf:"key-hex"`         // IMGPROXY_KEY_HEX
	SaltHex        string `koanf:"salt-hex"`        // IMGPROXY_SALT_HEX
	DefaultQuality int
	Key            []byte
	Salt           []byte
}

func (c *Config) ApplyDefaults() {
	c.PublicBaseURL = strings.TrimRight(c.PublicBaseURL, "/")
	c.Key, _ = hex.DecodeString(c.KeyHex)
	c.Salt, _ = hex.DecodeString(c.SaltHex)
	c.DefaultQuality = 80
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
	if c.DefaultQuality < 1 || c.DefaultQuality > 100 {
		return errors.New("imgproxy default quality must be between 1 and 100")
	}
	if c.Bucket == "" {
		return errors.New("imgproxy bucket is required")
	}
	return nil
}
