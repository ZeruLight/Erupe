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
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:      config.Logger,
		db:          config.DB,
		erupeConfig: config.ErupeConfig,
		acceptConns: make(chan net.Conn),
		deleteConns: make(chan net.Conn),
		sessions:    make(map[net.Conn]*Session),
		stages:      make(map[string]*Stage),
	}

	// Default town stage that clients try to enter without creating.
	stage := NewStage("sl1Ns200p0a0u0")
	s.stages[stage.id] = stage

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
