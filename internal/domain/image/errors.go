package image //nolint:revive // package name intentional

import "errors"

var (
	ErrImageNotFound       = errors.New("image not found")
	ErrImageTooLarge       = errors.New("image too large")
	ErrInvalidSize         = errors.New("invalid size")
	ErrUnsupportedMimeType = errors.New("unsupported mime type")
)
