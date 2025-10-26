package imgproxy

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	BaseURL        string `mapstructure:"base-url"` // IMGPROXY_BASE_URL
	KeyHex         string `mapstructure:"key-hex"`  // IMGPROXY_KEY_HEX
	SaltHex        string `mapstructure:"salt-hex"` // IMGPROXY_SALT_HEX
	DefaultQuality int
	Key            []byte
	Salt           []byte
}

func newConfig(v *viper.Viper) (Config, error) {
	var cfg Config
	if err := v.Sub("imgproxy").UnmarshalExact(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to load imgproxy config: %w", err)
	}
	if cfg.BaseURL == "" {
		return cfg, errors.New("imgproxy base URL is required")
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	key, err := hex.DecodeString(cfg.KeyHex)
	if err != nil {
		return cfg, fmt.Errorf("failed to decode key: %w", err)
	}
	cfg.Key = key
	salt, err := hex.DecodeString(cfg.SaltHex)
	if err != nil {
		return cfg, fmt.Errorf("failed to decode salt: %w", err)
	}
	cfg.Salt = salt
	cfg.DefaultQuality = 80

	return cfg, nil
}
