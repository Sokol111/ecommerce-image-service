package s3

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-image-service/internal/application/image"
	"go.uber.org/zap"
)

type imageTenantCleaner struct {
	storage image.ObjectStorage
	log     *zap.Logger
}

func NewImageTenantCleaner(storage image.ObjectStorage, log *zap.Logger) *imageTenantCleaner {
	return &imageTenantCleaner{storage: storage, log: log}
}

func (c *imageTenantCleaner) CleanupTenant(ctx context.Context, slug string) error {
	prefix := slug + "/"

	c.log.Info("deleting tenant images", zap.String("tenant", slug), zap.String("prefix", prefix))

	if err := c.storage.DeleteByPrefix(ctx, prefix); err != nil {
		return fmt.Errorf("failed to delete images for tenant %q: %w", slug, err)
	}

	c.log.Info("tenant images deleted", zap.String("tenant", slug))
	return nil
}
