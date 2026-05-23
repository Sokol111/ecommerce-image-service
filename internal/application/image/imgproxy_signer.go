package image //nolint:revive // package name intentional

import "time"

// SignerOptions contains parameters for building image transformation URLs
type SignerOptions struct {
	Width   *int
	Height  *int
	Fit     *string // fit | fill | fill-down | force | auto
	Quality *int    // 1..100
	DPR     *float32
	Format  *string    // webp | avif | jpeg | png | "" (original)
	Expires *time.Time // expiration time for signed URLs
}

// ImgproxySigner builds signed URLs for image transformation service
type ImgproxySigner interface {
	BuildURL(key string, opts SignerOptions) string
}
