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

type CreatePresignDTO struct {
	ContentType string
	Filename    string
	OwnerType   string
	OwnerId     string
	Role        string
	Size        int64
}

type CreatePresignResponseDTO struct {
	UploadUrl       string
	Key             string
	ExpiresIn       int
	RequiredHeaders map[string]string
}

type ConfirmUploadDTO struct {
	Alt       string
	Checksum  *string
	Key       string
	Mime      string
	OwnerType string
	OwnerId   string
	Role      string
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
	CreatePresign(ctx context.Context, dto CreatePresignDTO) (CreatePresignResponseDTO, error)

	ConfirmUpload(ctx context.Context, dto ConfirmUploadDTO) (*Image, error)

	PromoteDraftImages(ctx context.Context, dto PromoteDraftDTO) ([]*Image, error)

	GetDeliveryUrl(ctx context.Context, opts GetDeliveryUrlDTO) (string, *time.Time, error)

	DeleteImage(ctx context.Context, imageId string, hard bool) error

	GetImageById(ctx context.Context, imageId string) (*Image, error)
}
