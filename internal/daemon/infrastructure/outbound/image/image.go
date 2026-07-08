package image

import (
	domain "boyler/internal/daemon/core"
	"context"
)

type ImageManager interface {
	// Extract unpack .tar.gz archive
	Extract(ctx context.Context, name string, unpackDir string) error

	// IsExtracted check if image is extracted
	IsExtracted(ctx context.Context, name string) bool

	// GetRootfsPath return directory of rootfs
	GetRootfsPath(name string) string

	// Delete remove image
	Delete(ctx context.Context, name string) error

	// Get return mage using name
	Get(ctx context.Context, name string) (*domain.Image, error)

	// List return list of images
	List(ctx context.Context) ([]*domain.Image, error)

	// Pull download images from dockerHub
	Pull(ctx context.Context, name string) (*domain.Image, error)
}