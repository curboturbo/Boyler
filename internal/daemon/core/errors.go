package core

import "fmt"

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
	Err error
}

func (e *InternalDaemonError) Error() string { return fmt.Sprintf("daemon internal error :%v", e.Err) }
func (e *InternalDaemonError) Unwrap() error { return e.Err }