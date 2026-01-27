package apperrors

import "errors"

// Application layer errors for infrastructure/integration operations.
// For domain/business rule errors, use domain/image/errors.go

var (
	// ErrInvalidUploadToken indicates the upload token is invalid or expired (JWT validation)
	ErrInvalidUploadToken = errors.New("invalid upload token")

	// ErrObjectNotFound indicates the S3 object was not found (storage infrastructure)
	ErrObjectNotFound = errors.New("object not found in storage")

	// ErrDraftNotFound indicates the draft was not found (use-case specific)
	ErrDraftNotFound = errors.New("draft not found")

	// ErrInvalidImageOwner indicates the image has invalid owner for the operation (use-case validation)
	ErrInvalidImageOwner = errors.New("invalid image owner")
)
