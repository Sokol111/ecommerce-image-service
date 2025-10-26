package image

import "errors"

var (
	ErrImageNotFound       = errors.New("image not found")
	ErrInvalidOwnerType    = errors.New("invalid owner type")
	ErrInvalidImageKey     = errors.New("invalid image key")
	ErrImageTooLarge       = errors.New("image too large")
	ErrUnsupportedMimeType = errors.New("unsupported mime type")
	ErrImageAlreadyDeleted = errors.New("image already deleted")
	ErrCannotPromoteDraft  = errors.New("only draft images can be promoted")
)
