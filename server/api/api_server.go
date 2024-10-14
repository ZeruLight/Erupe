package api

import (
	"context"
	"erupe-ce/config"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"erupe-ce/utils/logger"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// APIServer is Erupes Standard API interface
type APIServer struct {
	sync.Mutex
	logger         logger.Logger
	erupeConfig    *config.Config
	httpServer     *http.Server
	isShuttingDown bool
}

// NewAPIServer creates a new Server type.
func NewAPIServer() *APIServer {

	s := &APIServer{
		logger:     logger.Get().Named("api"),
		httpServer: &http.Server{},
	}
	return s
}

// Start starts the server in a new goroutine.
func (s *APIServer) Start() error {
	// Set up the routes responsible for serving the launcher HTML, serverlist, unique name check, and JP auth.
	r := mux.NewRouter()
	r.HandleFunc("/launcher", s.Launcher)
	r.HandleFunc("/login", s.Login)
	r.HandleFunc("/register", s.Register)
	r.HandleFunc("/character/create", s.CreateCharacter)
	r.HandleFunc("/character/delete", s.DeleteCharacter)
	r.HandleFunc("/character/export", s.ExportSave)
	r.HandleFunc("/api/ss/bbs/upload.php", s.ScreenShot)
	r.HandleFunc("/api/ss/bbs/{id}", s.ScreenShotGet)
	handler := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))(r)
	s.httpServer.Handler = handlers.LoggingHandler(os.Stdout, handler)
	s.httpServer.Addr = fmt.Sprintf(":%d", config.GetConfig().API.Port)

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
func (s *APIServer) Shutdown() {
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
