package imgproxy

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/knadh/koanf/v2"
)

type Config struct {
	PublicBaseURL  string `koanf:"public-base-url"` // IMGPROXY_PUBLIC_BASE_URL
	KeyHex         string `koanf:"key-hex"`         // IMGPROXY_KEY_HEX
	SaltHex        string `koanf:"salt-hex"`        // IMGPROXY_SALT_HEX
	DefaultQuality int
	Key            []byte
	Salt           []byte
}

func newConfig(k *koanf.Koanf) (Config, error) {
	var cfg Config
	if err := k.Unmarshal("imgproxy", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load imgproxy config: %w", err)
	}
	if cfg.PublicBaseURL == "" {
		return cfg, errors.New("imgproxy public base URL is required")
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")

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
