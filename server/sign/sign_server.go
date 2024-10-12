package sign

import (
	"fmt"
	"io"
	"net"
	"sync"

	"erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/utils/logger"

	"go.uber.org/zap"
)

// SignServer is a MHF sign server.
type SignServer struct {
	sync.Mutex
	logger         logger.Logger
	sessions       map[int]*Session
	listener       net.Listener
	isShuttingDown bool
}

// NewServer creates a new Server type.
func NewServer() *SignServer {
	s := &SignServer{
		logger: logger.Get().Named("sign"),
	}
	return s
}

// Start starts the server in a new goroutine.
func (server *SignServer) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GetConfig().Sign.Port))
	if err != nil {
		return err
	}
	server.listener = l

	go server.acceptClients()

	return nil
}

// Shutdown exits the server gracefully.
func (server *SignServer) Shutdown() {
	server.logger.Debug("Shutting down...")

	server.Lock()
	server.isShuttingDown = true
	server.Unlock()

	// This will cause the acceptor goroutine to error and exit gracefully.
	server.listener.Close()
}

func (server *SignServer) acceptClients() {
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

func (server *SignServer) handleConnection(conn net.Conn) {
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
