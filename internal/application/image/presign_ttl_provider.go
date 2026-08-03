package image

import "time"

// PresignTTLProvider provides the presign TTL for upload token generation
type PresignTTLProvider interface {
	GetPresignTTL() time.Duration
}
