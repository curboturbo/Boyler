package image

import "boyler/internal/daemon/domain"

type ImageManager interface {
	// Extract unpack .tar.gz archive
	Extract(name string, unpackDir string) error

	// IsExtracted check if image is extracted
	IsExtracted(name string) bool

	// GetRootfsPath return directory of rootfs
	GetRootfsPath(name string) string

	// Delete remove image
	Delete(name string) error

	// Get return mage using name
	Get(name string) (*domain.Image, error)

	// List return list of images
	List() ([]*domain.Image, error)

	// Pull download images from dockerHub
	Pull(name string) (*domain.Image, error)
}