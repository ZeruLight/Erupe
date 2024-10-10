package signserver

import (
	"fmt"
	"io"
	"net"
	"sync"

	_config "erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/utils/logger"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	DB          *sqlx.DB
	ErupeConfig *_config.Config
}

// Server is a MHF sign server.
type Server struct {
	sync.Mutex
	logger         logger.Logger
	erupeConfig    *_config.Config
	sessions       map[int]*Session
	db             *sqlx.DB
	listener       net.Listener
	isShuttingDown bool
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:      logger.Get().Named("sign"),
		erupeConfig: config.ErupeConfig,
		db:          config.DB,
	}
	return s
}

// Start starts the server in a new goroutine.
func (server *Server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", server.erupeConfig.Sign.Port))
	if err != nil {
		return err
	}
	server.listener = l

	go server.acceptClients()

	return nil
}

// Shutdown exits the server gracefully.
func (server *Server) Shutdown() {
	server.logger.Debug("Shutting down...")

	server.Lock()
	server.isShuttingDown = true
	server.Unlock()

	// This will cause the acceptor goroutine to error and exit gracefully.
	server.listener.Close()
}

func (server *Server) acceptClients() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			// Check if we are shutting down and exit gracefully if so.
			server.Lock()
			shutdown := server.isShuttingDown
			server.Unlock()

			if shutdown {
				break
			} else {
				panic(err)
			}
		}

		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	server.logger.Debug("New connection", zap.String("RemoteAddr", conn.RemoteAddr().String()))
	defer conn.Close()

	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	_, err := io.ReadFull(conn, nullInit)
	if err != nil {
		server.logger.Error("Error initializing connection", zap.Error(err))
		return
	}

	// Create a new session.
	session := &Session{
		logger:    server.logger,
		server:    server,
		rawConn:   conn,
		cryptConn: network.NewCryptConn(conn),
	}

	// Do the session's work.
	session.work()
}
