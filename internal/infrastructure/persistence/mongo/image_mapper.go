package mongo

import (
	"github.com/Sokol111/ecommerce-image-service/internal/domain/image"
)

type imageMapper struct{}

func newImageMapper() *imageMapper {
	return &imageMapper{}
}

func (m *imageMapper) ToEntity(img *image.Image) *imageEntity {
	return &imageEntity{
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

func (m *imageMapper) ToDomain(e *imageEntity) *image.Image {
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

func (m *imageMapper) GetID(e *imageEntity) string {
	return e.ID
}

func (m *imageMapper) GetVersion(e *imageEntity) int {
	return e.Version
}

func (m *imageMapper) SetVersion(e *imageEntity, version int) {
	e.Version = version
}
