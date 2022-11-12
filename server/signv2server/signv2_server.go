package signv2server

import (
	"context"
	"erupe-ce/config"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Config struct {
	Logger      *zap.Logger
	DB          *sqlx.DB
	ErupeConfig *config.Config
}

// Server is the MHF custom launcher sign server.
type Server struct {
	sync.Mutex
	logger         *zap.Logger
	erupeConfig    *config.Config
	db             *sqlx.DB
	httpServer     *http.Server
	isShuttingDown bool
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		logger:      config.Logger,
		erupeConfig: config.ErupeConfig,
		db:          config.DB,
		httpServer:  &http.Server{},
	}
	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	// Set up the routes responsible for serving the launcher HTML, serverlist, unique name check, and JP auth.
	r := mux.NewRouter()
	r.HandleFunc("/launcher", s.Launcher)
	r.HandleFunc("/login", s.Login)
	r.HandleFunc("/register", s.Register)
	r.HandleFunc("/character/create", s.CreateCharacter)
	r.HandleFunc("/character/delete", s.DeleteCharacter)
	handler := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))(r)
	s.httpServer.Handler = handlers.LoggingHandler(os.Stdout, handler)
	s.httpServer.Addr = fmt.Sprintf(":%d", s.erupeConfig.SignV2.Port)

	serveError := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			// Send error if any.
			serveError <- err
		}
	}()

	// Get the error from calling ListenAndServe, otherwise assume it's good after 250 milliseconds.
	select {
	case err := <-serveError:
		return err
	case <-time.After(250 * time.Millisecond):
		return nil
	}
}

// Shutdown exits the server gracefully.
func (s *Server) Shutdown() {
	s.logger.Debug("Shutting down")

	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		// Just warn because we are shutting down the server anyway.
		s.logger.Warn("Got error on httpServer shutdown", zap.Error(err))
	}
}
