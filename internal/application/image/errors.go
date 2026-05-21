package image //nolint:revive // package name intentional

import "errors"

var (
	ErrImageNotFound       = errors.New("image not found")
	ErrImageTooLarge       = errors.New("image too large")
	ErrInvalidSize         = errors.New("invalid size")
	ErrUnsupportedMimeType = errors.New("unsupported mime type")
	ErrInvalidUploadToken  = errors.New("invalid upload token")
	ErrObjectNotFound      = errors.New("object not found in storage")
	ErrDraftNotFound       = errors.New("draft not found")
	ErrInvalidImageOwner   = errors.New("invalid image owner")
)
