package core

import "time"

type Image struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Tag        string      `json:"tag"`
	Size       int64       `json:"size"`
	CreatedAt  time.Time   `json:"created_at"`
	RootfsPath string      `json:"rootfs_path"`
	TarPath    string      `json:"tar_path"`
	Config     ImageConfig `json:"config"`
}


type ImageConfig struct {
	Env          []string          `json:"env"`
	Cmd          []string          `json:"cmd"`
	Entrypoint   []string          `json:"entrypoint"`
	WorkingDir   string            `json:"working_dir"`
	User         string            `json:"user"`
	ExposedPorts map[string]struct{} `json:"exposed_ports"`
	Labels       map[string]string `json:"labels"`
}