package entranceserver

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	_config "erupe-ce/config"
	"erupe-ce/network"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Server is a MHF entrance server.
type Server struct {
	sync.Mutex
	logger         *zap.Logger
	erupeConfig    *_config.Config
	db             *sqlx.DB
	listener       net.Listener
	isShuttingDown bool
}

// Config struct allows configuring the server.
type Config struct {
	Logger      *zap.Logger
	DB          *sqlx.DB
	ErupeConfig *_config.Config
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	server := &Server{
		logger:      config.Logger,
		erupeConfig: config.ErupeConfig,
		db:          config.DB,
	}
	return server
}

// Start starts the server in a new goroutine.
func (server *Server) Start() error {

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", server.erupeConfig.Entrance.Port))
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

// acceptClients handles accepting new clients in a loop.
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
				continue
			}
		}

		// Start a new goroutine for the connection so that we don't block other incoming connections.
		go server.handleEntranceServerConnection(conn)
	}
}

func (server *Server) handleEntranceServerConnection(conn net.Conn) {
	defer conn.Close()
	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	n, err := io.ReadFull(conn, nullInit)
	if err != nil {
		server.logger.Warn("Failed to read 8 NULL init", zap.Error(err))
		return
	} else if n != len(nullInit) {
		server.logger.Warn("io.ReadFull couldn't read the full 8 byte init.")
		return
	}

	// Create a new encrypted connection handler and read a packet from it.
	cc := network.NewCryptConn(conn)
	pkt, err := cc.ReadPacket()
	if err != nil {
		server.logger.Warn("Error reading packet", zap.Error(err))
		return
	}

	if server.erupeConfig.DebugOptions.LogInboundMessages {
		fmt.Printf("[Client] -> [Server]\nData [%d bytes]:\n%s\n", len(pkt), hex.Dump(pkt))
	}

	local := false
	if strings.Split(conn.RemoteAddr().String(), ":")[0] == "127.0.0.1" {
		local = true
	}
	data := makeSv2Resp(server.erupeConfig, server, local)
	if len(pkt) > 5 {
		data = append(data, makeUsrResp(pkt, server)...)
	}
	cc.SendPacket(data)
	// Close because we only need to send the response once.
	// Any further requests from the client will come from a new connection.
}
