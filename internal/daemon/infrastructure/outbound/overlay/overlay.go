package overlay

import (
	"context"
)


type VolumeManager interface {
	// CreateMountPoints create upperdir, workdir and merged for each container
	CreateMountPoints(ctx context.Context, containerID string) error

	// Mount build layers (lowerdir from image + upperdir) to merged
	Mount(ctx context.Context, containerID string, imageName string) error

	// Unmount remount merged before delete
	Unmount(ctx context.Context, containerID string) error

	// Cleanup delete upperdir, workdir, merged
	Cleanup(ctx context.Context, containerID string) error
}