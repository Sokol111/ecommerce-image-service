package image

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Sokol111/ecommerce-image-service/internal/model"
	"github.com/Sokol111/ecommerce-image-service/internal/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
)

type service struct {
	store      Store
	bucket     string
	client     *awss3.Client
	presigner  *awss3.PresignClient
	presignTTL time.Duration
	maxBytes   int64
	// outbox    outbox.Outbox
	// txManager mongo.TxManager
}

func NewService(store Store, cfg s3.Config, c *awss3.Client, p *awss3.PresignClient) model.ImageService {
	return &service{
		store:      store,
		bucket:     cfg.Bucket,
		client:     c,
		presigner:  p,
		presignTTL: cfg.PresignTTL,
		maxBytes:   cfg.MaxUploadBytes,
	}
}

func (s *service) CreateDraftPresign(ctx context.Context, dto model.CreatePresignForDraftDTO) (model.CreatePresignForDraftResponseDTO, error) {
	extByCT := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/avif": ".avif",
	}
	ext := extByCT[strings.ToLower(dto.ContentType)]
	if ext == "" {
		return model.CreatePresignForDraftResponseDTO{}, fmt.Errorf("unsupported content type: %s", dto.ContentType)
	}

	key := "product-drafts/" + dto.DraftId + "/" + uuid.New().String() + "." + ext

	out, err := s.presigner.PresignPutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(string(dto.ContentType)),
	}, awss3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return model.CreatePresignForDraftResponseDTO{}, fmt.Errorf("presign put: %w", err)
	}

	return model.CreatePresignForDraftResponseDTO{
		UploadUrl: out.URL,
		Key:       key,
		ExpiresIn: int(s.presignTTL.Seconds()),
		RequiredHeaders: map[string]string{
			"Content-Type": string(dto.ContentType),
		},
	}, nil
}

func (s *service) ConfirmDraftUpload(ctx context.Context, dto model.ConfirmDraftUploadDTO) (*model.Image, error) {
	expectedPrefix := "product-drafts/" + dto.DraftId + "/"

	if !strings.HasPrefix(dto.Key, expectedPrefix) {
		return nil, fmt.Errorf("key does not match expected owner prefix")
	}

	ho, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(dto.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("head object: %w", err)
	}

	size := int64(0)
	if ho.ContentLength != nil {
		size = *ho.ContentLength
	}
	if s.maxBytes > 0 && size > s.maxBytes {
		_, _ = s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(dto.Key),
		})
		return nil, fmt.Errorf("File too large: max %d bytes", s.maxBytes)
	}

	now := time.Now().UTC()

	image := &model.Image{
		Id:         uuid.New().String(),
		Alt:        dto.Alt,
		OwnerType:  "productDraft",
		OwnerId:    dto.DraftId,
		Role:       dto.Role,
		Key:        dto.Key,
		Mime:       dto.Mime,
		Size:       size,
		Status:     "uploaded",
		CreatedAt:  now,
		ModifiedAt: now,
		Version:    1,
	}
	created, err := s.store.Create(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("create image doc: %w", err)
	}
	return created, nil
}

func (s *service) PromoteDraftImages(ctx context.Context, dto model.PromoteDraftDTO) ([]*model.Image, error) {
	var wantIDs []string
	if dto.Images != nil && len(*dto.Images) > 0 {
		wantIDs = *dto.Images
	}

	imgs, err := s.store.ListByOwner(ctx, "productDraft", dto.DraftId, wantIDs)
	if err != nil {
		return nil, fmt.Errorf("list draft images: %w", err)
	}
	if len(imgs) == 0 {
		return []*model.Image{}, fmt.Errorf("no images found for draft %s", dto.DraftId)
	}

	var promoted []*model.Image
	srcPrefix := "product-drafts/" + dto.DraftId + "/"
	for _, im := range imgs {
		if !strings.HasPrefix(im.Key, srcPrefix) {
			return nil, fmt.Errorf("image %s has key outside draft prefix: %s", im.Id, im.Key)
		}

		dstKey := "products/" + dto.ProductId + "/" + strings.TrimPrefix(im.Key, srcPrefix)

		exists, err := s.objectExists(ctx, dstKey)
		if err != nil {
			return nil, fmt.Errorf("check target exists: %w", err)
		}
		if !exists {
			copySource := url.PathEscape(s.bucket + "/" + im.Key)
			_, err = s.client.CopyObject(ctx, &awss3.CopyObjectInput{
				Bucket:     aws.String(s.bucket),
				Key:        aws.String(dstKey),
				CopySource: aws.String(copySource),
			})
			if err != nil {
				return nil, fmt.Errorf("copy %s -> %s: %w", im.Key, dstKey, err)
			}
		}

		_, _ = s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(im.Key),
		})

		updated, err := s.store.UpdateAfterPromote(ctx, im.Id, "product", dto.ProductId, dstKey)
		if err != nil {
			return nil, fmt.Errorf("update db after promote: %w", err)
		}

		promoted = append(promoted, updated)
	}
	return promoted, nil
}

func (s *service) objectExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isS3NotFound(err error) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if errors.As(err, &ae) {
		switch ae.ErrorCode() {
		case "NotFound", "NoSuchKey", "NoSuchBucket":
			return true
		}
	}
	var re *awshttp.ResponseError
	if errors.As(err, &re) && re.HTTPStatusCode() == 404 {
		return true
	}
	return false
}
