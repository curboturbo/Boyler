package core

import "time"

type Image struct {
	ID         string
	Name       string
	Tag        string
	Size       int64
	CreatedAt  time.Time
	RootfsPath string
	TarPath    string
}
