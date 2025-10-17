package model

import (
	"context"
	"time"
)

type Image struct {
	Id         string
	Version    int
	Alt        string
	OwnerType  string
	OwnerId    string
	Role       string
	Key        string
	Mime       string
	Size       int64
	Status     string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type CreatePresignForDraftDTO struct {
	ContentType string
	Filename    string
	DraftId     string
	Size        int64
}

type CreatePresignForDraftResponseDTO struct {
	UploadUrl       string
	Key             string
	ExpiresIn       int
	RequiredHeaders map[string]string
}

type ConfirmDraftUploadDTO struct {
	Alt      string
	Checksum *string
	Key      string
	Mime     string
	DraftId  string
	Role     string
}

type PromoteDraftDTO struct {
	DraftId   string
	Images    *[]string
	Move      bool
	ProductId string
}

type ImageService interface {
	CreateDraftPresign(ctx context.Context, dto CreatePresignForDraftDTO) (CreatePresignForDraftResponseDTO, error)

	ConfirmDraftUpload(ctx context.Context, dto ConfirmDraftUploadDTO) (*Image, error)

	PromoteDraftImages(ctx context.Context, dto PromoteDraftDTO) ([]*Image, error)
}
