package image

import (
	"fmt"
)

type Config struct {
	// MaxUploadBytes is the maximum allowed file upload size in bytes
	MaxUploadBytes int64 `koanf:"max-upload-bytes"`
	// SmallWidth is the width for small product images
	SmallWidth int `koanf:"small-width"`
	// LargeWidth is the width for large product images
	LargeWidth int `koanf:"large-width"`
	// Quality is the JPEG quality for product images (1-100)
	Quality int `koanf:"quality"`
}

func (c *Config) ApplyDefaults() {
	if c.MaxUploadBytes == 0 {
		c.MaxUploadBytes = 5 * 1024 * 1024 // 5 MB default
	}
	if c.SmallWidth == 0 {
		c.SmallWidth = 400
	}
	if c.LargeWidth == 0 {
		c.LargeWidth = 800
	}
	if c.Quality == 0 {
		c.Quality = 85
	}
}

func (c *Config) Validate() error {
	if c.MaxUploadBytes <= 0 {
		return fmt.Errorf("max-upload-bytes must be greater than 0")
	}
	if c.MaxUploadBytes > 10*1024*1024 {
		return fmt.Errorf("max-upload-bytes must not exceed 10 MB")
	}
	if c.SmallWidth <= 0 {
		return fmt.Errorf("small-width must be greater than 0")
	}
	if c.LargeWidth <= 0 {
		return fmt.Errorf("large-width must be greater than 0")
	}
	if c.Quality < 1 || c.Quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100")
	}
	return nil
}
