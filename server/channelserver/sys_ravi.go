package channelserver

import (
	"strings"
	"sync"
)

type Raviente struct {
	sync.Mutex
	id       uint16
	register []uint32
	state    []uint32
	support  []uint32
}

func (server *Server) resetRaviente() {
	for _, semaphore := range server.semaphore {
		if strings.HasPrefix(semaphore.name, "hs_l0") {
			return
		}
	}
	server.logger.Debug("All Raviente Semaphores empty, resetting")
	server.raviente.id = server.raviente.id + 1
	server.raviente.register = make([]uint32, 30)
	server.raviente.state = make([]uint32, 30)
	server.raviente.support = make([]uint32, 30)
}

func (server *Server) GetRaviMultiplier() float64 {
	raviSema := server.getRaviSemaphore()
	if raviSema != nil {
		var minPlayers int
		if server.raviente.register[9] > 8 {
			minPlayers = 24
		} else {
			minPlayers = 4
		}
		if len(raviSema.clients) > minPlayers {
			return 1
		}
		return float64(minPlayers / len(raviSema.clients))
	}
	return 0
}

func (server *Server) UpdateRavi(semaID uint32, index uint8, value uint32, update bool) (uint32, uint32) {
	var prev uint32
	var dest *[]uint32
	switch semaID {
	case 0x40000:
		switch index {
		case 17, 28: // Ignore res and poison
			break
		default:
			value = uint32(float64(value) * server.GetRaviMultiplier())
		}
		dest = &server.raviente.state
	case 0x50000:
		dest = &server.raviente.support
	case 0x60000:
		dest = &server.raviente.register
	default:
		return 0, 0
	}
	if update {
		(*dest)[index] += value
	} else {
		(*dest)[index] = value
	}
	return prev, (*dest)[index]
}
