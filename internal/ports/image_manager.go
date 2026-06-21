package ports

import "boyler/internal/domain"

type ImageManager interface {
	// Pull downloads an image from a registry
	Pull(name string) (*domain.Image, error)

	// List returns all locally stored images
	List() ([]*domain.Image, error)
	
	// Get retrieves an image by name
	Get(name string) (*domain.Image, error)
	
	// Delete removes an image by ID
	Delete(id string) error

	// Extract unpacks an image archive to filesystem
	Extract(name string, unpackDir string) error

	// IsExtracted checks if an image is unpacked
	IsExtracted(name string) bool

	// GetRootfsPath returns the root filesystem path
	GetRootfsPath(name string) string
}