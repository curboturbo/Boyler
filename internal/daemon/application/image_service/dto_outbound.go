package imageservice

type Event struct{
	Status string
	LayId string
	Progress int64
	Total int64
}