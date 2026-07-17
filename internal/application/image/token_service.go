package image //nolint:revive // package name intentional

import (
	"context"
	"time"
)

// UploadTokenClaims represents the JWT claims for presigned upload
type UploadTokenClaims struct {
	Key         string `json:"key"`
	Tenant      string `json:"tenant"`
	OwnerType   string `json:"ownerType"`
	OwnerID     string `json:"ownerId"`
	Role        string `json:"role"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// TokenService handles JWT token generation and validation for uploads
type TokenService interface {
	// GenerateUploadToken creates a signed JWT token with upload metadata
	GenerateUploadToken(ctx context.Context, claims *UploadTokenClaims, ttl time.Duration) (string, error)

	// ValidateUploadToken verifies and parses a JWT token, returning the claims
	ValidateUploadToken(ctx context.Context, token string) (*UploadTokenClaims, error)
}
