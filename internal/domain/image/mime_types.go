package image

import "strings"

// SupportedMimeTypes defines allowed image MIME types and their file extensions.
// This is the single source of truth for content type validation.
var SupportedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/avif": ".avif",
}

// GetExtensionForMimeType returns the file extension for a given MIME type.
// Returns empty string if the MIME type is not supported.
func GetExtensionForMimeType(mimeType string) string {
	return SupportedMimeTypes[strings.ToLower(mimeType)]
}

// IsSupportedMimeType checks if the given MIME type is supported.
func IsSupportedMimeType(mimeType string) bool {
	_, ok := SupportedMimeTypes[strings.ToLower(mimeType)]
	return ok
}
