package model

import (
	"context"
	"errors"
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

type GetDeliveryUrlDTO struct {
	ImageId string
	Width   *int
	Height  *int
	Fit     *string // fit | fill | fill-down | force | auto
	Quality *int    // 1..100
	DPR     *float32
	Format  *string    // webp | avif | jpeg | png | "" (оригінал)
	Expires *time.Time // якщо хочеш “exp:unix”
}

var ErrEntityNotFound = errors.New("entity not found")

type ImageService interface {
	CreateDraftPresign(ctx context.Context, dto CreatePresignForDraftDTO) (CreatePresignForDraftResponseDTO, error)

	ConfirmDraftUpload(ctx context.Context, dto ConfirmDraftUploadDTO) (*Image, error)

	PromoteDraftImages(ctx context.Context, dto PromoteDraftDTO) ([]*Image, error)

	GetDeliveryUrl(ctx context.Context, opts GetDeliveryUrlDTO) (string, *time.Time, error)
}
