package mongo

import (
	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
)

type ImageMapper struct{}

func NewImageMapper() *ImageMapper {
	return &ImageMapper{}
}

func (m *ImageMapper) ToEntity(img *image.Image) *ImageEntity {
	return &ImageEntity{
		ID:         img.ID,
		Version:    img.Version,
		Alt:        img.Alt,
		OwnerType:  img.OwnerType,
		OwnerID:    img.OwnerID,
		Role:       img.Role,
		Key:        img.Key,
		Mime:       img.Mime,
		Size:       img.Size,
		Status:     string(img.Status),
		CreatedAt:  img.CreatedAt,
		ModifiedAt: img.ModifiedAt,
	}
}

func (m *ImageMapper) ToDomain(e *ImageEntity) *image.Image {
	return image.Reconstruct(
		e.ID,
		e.Version,
		e.Alt,
		e.OwnerType,
		e.OwnerID,
		e.Role,
		e.Key,
		e.Mime,
		e.Size,
		image.ImageStatus(e.Status),
		e.CreatedAt.UTC(),
		e.ModifiedAt.UTC(),
	)
}

func (m *ImageMapper) GetID(e *ImageEntity) string {
	return e.ID
}

func (m *ImageMapper) GetVersion(e *ImageEntity) int64 {
	return e.Version
}

func (m *ImageMapper) SetVersion(e *ImageEntity, version int64) {
	e.Version = version
}
