package image

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Image - domain aggregate root
type Image struct {
	ID         string
	Version    int
	Alt        string
	OwnerType  string
	OwnerID    string
	Role       string
	Key        string
	Mime       string
	Size       int64
	Status     ImageStatus
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type ImageStatus string

const (
	StatusUploaded   ImageStatus = "uploaded"
	StatusProcessing ImageStatus = "processing"
	StatusReady      ImageStatus = "ready"
	StatusDeleted    ImageStatus = "deleted"
)

// NewImage creates a new image with validation
func NewImage(alt, ownerType, ownerID, role, key, mime string, size int64) (*Image, error) {
	if err := validateImageData(ownerType, ownerID, key, mime, size); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Image{
		ID:         uuid.New().String(),
		Version:    1,
		Alt:        alt,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		Role:       role,
		Key:        key,
		Mime:       mime,
		Size:       size,
		Status:     StatusUploaded,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

// NewImageWithID creates an image with a specific ID (for idempotency)
func NewImageWithID(id, alt, ownerType, ownerID, role, key, mime string, size int64) (*Image, error) {
	if err := validateImageData(ownerType, ownerID, key, mime, size); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Image{
		ID:         id,
		Version:    1,
		Alt:        alt,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		Role:       role,
		Key:        key,
		Mime:       mime,
		Size:       size,
		Status:     StatusUploaded,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

// Reconstruct rebuilds an image from persistence (no validation)
func Reconstruct(id string, version int, alt, ownerType, ownerID, role, key, mime string, size int64, status ImageStatus, createdAt, modifiedAt time.Time) *Image {
	return &Image{
		ID:         id,
		Version:    version,
		Alt:        alt,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		Role:       role,
		Key:        key,
		Mime:       mime,
		Size:       size,
		Status:     status,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
	}
}

// UpdateAlt updates the image alt text
func (i *Image) UpdateAlt(alt string) {
	i.Alt = alt
	i.ModifiedAt = time.Now().UTC()
}

// PromoteToProduct promotes the image from draft to product
func (i *Image) PromoteToProduct(productID, newKey string) error {
	if i.OwnerType != "productDraft" {
		return errors.New("only draft images can be promoted")
	}

	i.OwnerType = "product"
	i.OwnerID = productID
	i.Key = newKey
	i.Status = StatusProcessing
	i.ModifiedAt = time.Now().UTC()
	return nil
}

// MarkAsReady marks the image as ready to use
func (i *Image) MarkAsReady() {
	i.Status = StatusReady
	i.ModifiedAt = time.Now().UTC()
}

// MarkAsProcessing marks the image as processing
func (i *Image) MarkAsProcessing() {
	i.Status = StatusProcessing
	i.ModifiedAt = time.Now().UTC()
}

// MarkAsDeleted soft deletes the image
func (i *Image) MarkAsDeleted() {
	i.Status = StatusDeleted
	i.ModifiedAt = time.Now().UTC()
}

// IncrementVersion increments version for optimistic locking
func (i *Image) IncrementVersion() {
	i.Version++
}

// IsDeleted checks if the image is marked as deleted
func (i *Image) IsDeleted() bool {
	return i.Status == StatusDeleted
}

// validateImageData validates business rules
func validateImageData(ownerType, ownerID, key, mime string, size int64) error {
	if ownerType == "" {
		return errors.New("owner type is required")
	}

	if ownerID == "" {
		return errors.New("owner ID is required")
	}

	if key == "" {
		return errors.New("key is required")
	}

	if mime == "" {
		return errors.New("mime type is required")
	}

	if size < 0 {
		return errors.New("size cannot be negative")
	}

	// Validate owner type
	switch ownerType {
	case "product", "productDraft", "user":
		// valid
	default:
		return errors.New("invalid owner type")
	}

	return nil
}
