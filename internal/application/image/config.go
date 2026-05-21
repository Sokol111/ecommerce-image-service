package image

// Config holds configuration for image handlers
type Config struct {
	// MaxUploadBytes is the maximum allowed file upload size in bytes
	MaxUploadBytes int64
	// SmallWidth is the width for small product images
	SmallWidth int
	// LargeWidth is the width for large product images
	LargeWidth int
	// Quality is the JPEG quality for product images (1-100)
	Quality int
}
