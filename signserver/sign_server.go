package signserver

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Andoryuuta/Erupe/network"
)

// Config struct allows configuring the server.
type Config struct {
	DB         *sql.DB
	ListenAddr string
}

// Server is a MHF sign server.
type Server struct {
	sync.Mutex
	sid        int
	sessions   map[int]*Session
	db         *sql.DB
	listenAddr string
	listener   net.Listener
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		sid:        0,
		sessions:   make(map[int]*Session),
		db:         config.DB,
		listenAddr: config.ListenAddr,
	}
	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	s.listener = l
	//defer l.Close()

	go s.acceptClients()

	return nil
}

func (s *Server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic(err)
		}

		go s.handleConnection(s.sid, conn)
		s.sid++
	}
}

func (s *Server) handleConnection(sid int, conn net.Conn) {
	// Client initalizes the connection with a one-time buffer of 8 NULL bytes.
	nullInit := make([]byte, 8)
	_, err := io.ReadFull(conn, nullInit)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	session := &Session{
		server:    s,
		rawConn:   &conn,
		cryptConn: network.NewCryptConn(conn),
	}

	s.Lock()
	s.sessions[sid] = session
	s.Unlock()

	session.work()
}
