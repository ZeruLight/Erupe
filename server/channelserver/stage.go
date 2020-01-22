package channelserver

import "sync"

// Stage holds stage-specific information
type Stage struct {
	sync.RWMutex
	gameObjectCount uint32
}
