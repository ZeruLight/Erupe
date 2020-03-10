package channelserver

import (
	"fmt"
	"net"
	"sync"

	"github.com/Andoryuuta/Erupe/config"
	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	Logger      *zap.Logger
	DB          *sqlx.DB
	ErupeConfig *config.Config
}

// Map key type for a user binary part.
type userBinaryPartID struct {
	charID uint32
	index  uint8
}

// Server is a MHF channel server.
type Server struct {
	sync.Mutex
	logger      *zap.Logger
	db          *sqlx.DB
	erupeConfig *config.Config
	acceptConns chan net.Conn
	deleteConns chan net.Conn
	sessions    map[net.Conn]*Session
	listener    net.Listener // Listener that is created when Server.Start is called.

	isShuttingDown bool

	stagesLock sync.RWMutex
	stages     map[string]*Stage

	userBinaryPartsLock sync.RWMutex
	userBinaryParts     map[userBinaryPartID][]byte
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:          config.Logger,
		db:              config.DB,
		erupeConfig:     config.ErupeConfig,
		acceptConns:     make(chan net.Conn),
		deleteConns:     make(chan net.Conn),
		sessions:        make(map[net.Conn]*Session),
		stages:          make(map[string]*Stage),
		userBinaryParts: make(map[userBinaryPartID][]byte),
	}

	// Default town stage that clients try to enter without creating.
	stage := NewStage("sl1Ns200p0a0u0")
	s.stages[stage.id] = stage

	// Town underground left area -- rasta bar stage (Maybe private bar ID as well?).
	stage2 := NewStage("sl1Ns211p0a0u0")
	s.stages[stage2.id] = stage2

	// Diva fountain / prayer fountain.
	stage3 := NewStage("sl2Ns379p0a0u0")
	s.stages[stage3.id] = stage3

	// sl1Ns257p0a0uE31111 -- house for charID E31111.

	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.erupeConfig.Channel.Port))
	if err != nil {
		return err
	}
	s.listener = l

	go s.acceptClients()
	go s.manageSessions()

	return nil
}

// Shutdown tries to shut down the server gracefully.
func (s *Server) Shutdown() {
	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	s.listener.Close()
	close(s.acceptConns)
}

func (s *Server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.Lock()
			shutdown := s.isShuttingDown
			s.Unlock()

			if shutdown {
				break
			} else {
				s.logger.Warn("Error accepting client", zap.Error(err))
				continue
			}
		}
		s.acceptConns <- conn
	}
}

func (s *Server) manageSessions() {
	for {
		select {
		case newConn := <-s.acceptConns:
			// Gracefully handle acceptConns channel closing.
			if newConn == nil {
				s.Lock()
				shutdown := s.isShuttingDown
				s.Unlock()

				if shutdown {
					return
				}
			}

			session := NewSession(s, newConn)

			s.Lock()
			s.sessions[newConn] = session
			s.Unlock()

			session.Start()

		case delConn := <-s.deleteConns:
			s.Lock()
			delete(s.sessions, delConn)
			s.Unlock()
		}
	}
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions.
func (s *Server) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	// Make the header
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(pkt.Opcode()))

	// Build the packet onto the byteframe.
	pkt.Build(bf)

	// Broadcast the data.
	for _, session := range s.sessions {
		if session == ignoredSession {
			continue
		}
		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}

func (s *Server) FindSessionByCharID(charID uint32) *Session {
	s.stagesLock.RLock()
	defer s.stagesLock.RUnlock()
	for _, stage := range s.stages {
		stage.RLock()
		for client := range stage.clients {
			if client.charID == charID {
				stage.RUnlock()
				return client
			}
		}
		stage.RUnlock()
	}

	return nil
}
