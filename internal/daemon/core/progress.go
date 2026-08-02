package core

type PullingEvent struct {
	Status string
	LayId string
	Progress int64
	Total int64
}
