package core

import (
	"fmt"
	"errors"
)

type ImageError struct {
	Image string
	Err   error
}

func (e *ImageError) Error() string { return fmt.Sprintf("image error (%s): %v", e.Image, e.Err) }
func (e *ImageError) Unwrap() error { return e.Err }

type FilesystemError struct {
	Op  string
	Err error
}

func (e *FilesystemError) Error() string { return fmt.Sprintf("filesystem error [%s]: %v", e.Op, e.Err) }
func (e *FilesystemError) Unwrap() error { return e.Err }

type RuntimeError struct {
	Op  string
	Err error
}

func (e *RuntimeError) Error() string { return fmt.Sprintf("runtime error [%s]: %v", e.Op, e.Err) }
func (e *RuntimeError) Unwrap() error { return e.Err }

type NetworkError struct {
	Op  string
	Err error
}

func (e *NetworkError) Error() string { return fmt.Sprintf("network error [%s]: %v", e.Op, e.Err) }
func (e *NetworkError) Unwrap() error { return e.Err }

type CgroupsError struct {
	Op  string
	Err error
}

func (e *CgroupsError) Error() string { return fmt.Sprintf("cgroups error [%s]: %v", e.Op, e.Err) }
func (e *CgroupsError) Unwrap() error { return e.Err }


type InternalDaemonError struct{
	Op string
	Err error
}

func (e *InternalDaemonError) Error() string { return fmt.Sprintf("daemon internal error [%s]: %v",e.Op, e.Err) }
func (e *InternalDaemonError) Unwrap() error { return e.Err }

type InvalidUserCommandError struct {
	Op string
	Err error
}

func (e *InvalidUserCommandError) Error() string { return fmt.Sprintf("daemon internal error [%s]: %v",e.Op, e.Err) }
func (e *InvalidUserCommandError) Unwrap() error { return e.Err }


/*
	General validation errors
*/

var (
	ErrInvalidRequest = errors.New("invalid request")

	ErrMissingRequiredField = errors.New("missing required field")

	ErrInvalidIdentifier = errors.New("invalid identifier")

	ErrInvalidName = errors.New("invalid name")

	ErrInvalidArgument = errors.New("invalid argument")

	ErrUnsupportedOption = errors.New("unsupported option")
)

/*
	Container creation validation
*/

var (
	ErrContainerNameRequired = errors.New("container name is required")

	ErrContainerNameTooLong = errors.New("container name too long")

	ErrContainerNameAlreadyUsed = errors.New("container name already used")

	ErrImageRequired = errors.New("image is required")

	ErrCommandRequired = errors.New("command is required")

	ErrInvalidCommand = errors.New("invalid command")

	ErrInvalidEnvironment = errors.New("invalid environment variable")

	ErrInvalidVolume = errors.New("invalid volume configuration")

	ErrInvalidPortMapping = errors.New("invalid port mapping")
)

/*
	Container lifecycle user errors
*/

var (
	ErrContainerNotFound = errors.New("container not found")

	ErrContainerAlreadyExists = errors.New("container already exists")

	ErrContainerAlreadyRunning = errors.New("container already running")

	ErrContainerNotRunning = errors.New("container is not running")

	ErrContainerPaused = errors.New("container is paused")

	ErrContainerNotPaused = errors.New("container is not paused")

	ErrContainerCannotStart = errors.New("container cannot be started")

	ErrContainerCannotStop = errors.New("container cannot be stopped")

	ErrContainerCannotRestart = errors.New("container cannot be restarted")

	ErrContainerCannotRemove = errors.New("container cannot be removed")
)

/*
	Image related user errors
*/

var (
	ErrImageNotFound = errors.New("image not found")

	ErrImageAlreadyExists = errors.New("image already exists")

	ErrInvalidImageReference = errors.New("invalid image reference")

	ErrUnsupportedImageFormat = errors.New("unsupported image format")

	ErrImagePullRequired = errors.New("image pull required")
)

/*
	Network user errors
*/

var (
	ErrNetworkNotFound = errors.New("network not found")

	ErrNetworkAlreadyExists = errors.New("network already exists")

	ErrInvalidNetworkConfig = errors.New("invalid network configuration")

	ErrPortAlreadyAllocated = errors.New("port already allocated")

	ErrInvalidIPAddress = errors.New("invalid ip address")

	ErrIpAddressNotExist = errors.New("container and ip addr not exist")
)

/*
	Resource limits validation
*/

var (
	ErrInvalidCPULimit = errors.New("invalid cpu limit")

	ErrInvalidMemoryLimit = errors.New("invalid memory limit")

	ErrInvalidResourceConfiguration = errors.New("invalid resource configuration")
)

/*
	Permission / access errors
*/

var (
	ErrPermissionDenied = errors.New("permission denied")

	ErrOperationNotAllowed = errors.New("operation not allowed")
)