package overlay

type VolumeManager interface {
	// CreateMountPoints create upperdir, workdir and merged for each container
	CreateMountPoints(containerID string) error

	// Mount build layers (lowerdir from image + upperdir) to merged
	Mount(lowerDir string, upperDir string, workDir string) error

	// Unmount remount merged before delete
	Unmount(containerID string) error

	// Cleanup delete upperdir, workdir, merged
	Cleanup(containerID string) error
}