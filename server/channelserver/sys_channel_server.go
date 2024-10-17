package channelserver

import (
	"fmt"
	"net"
	"sync"
	"time"

	"erupe-ce/config"
	"erupe-ce/internal/system"
	"erupe-ce/server/discordbot"
	"erupe-ce/utils/db"

	"erupe-ce/utils/gametime"
	"erupe-ce/utils/logger"

	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	ID         uint16
	DiscordBot *discordbot.DiscordBot
	Name       string
	Enable     bool
}

// Map key type for a user binary part.
type userBinaryPartID struct {
	charID uint32
	index  uint8
}

// ChannelServer is a MHF channel server.
type ChannelServer struct {
	sync.Mutex
	Channels       []*ChannelServer
	ID             uint16
	GlobalID       string
	IP             string
	Port           uint16
	logger         logger.Logger
	erupeConfig    *config.Config
	acceptConns    chan net.Conn
	deleteConns    chan net.Conn
	sessions       map[net.Conn]*Session
	objectIDs      map[*Session]uint16
	listener       net.Listener // Listener that is created when Server.Start is called.
	isShuttingDown bool

	stagesLock sync.RWMutex
	stages     map[string]*system.Stage

	// Used to map different languages
	i18n i18n

	// UserBinary
	userBinaryPartsLock sync.RWMutex
	userBinaryParts     map[userBinaryPartID][]byte

	// Semaphore
	semaphoreLock sync.RWMutex
	semaphore     map[string]*Semaphore

	// Discord chat integration
	discordBot *discordbot.DiscordBot

	name string

	raviente *Raviente

	questCacheData map[int][]byte
	questCacheTime map[int]time.Time
}

// NewServer creates a new Server type.
func NewServer(config *Config) *ChannelServer {
	stageNames := []string{
		"sl1Ns200p0a0u0", // Mezeporta
		"sl1Ns211p0a0u0", // Rasta bar
		"sl1Ns260p0a0u0", // Pallone Carvan
		"sl1Ns262p0a0u0", // Pallone Guest House 1st Floor
		"sl1Ns263p0a0u0", // Pallone Guest House 2nd Floor
		"sl2Ns379p0a0u0", // Diva fountain
		"sl1Ns462p0a0u0", // MezFes
	}
	stages := make(map[string]*system.Stage)
	for _, name := range stageNames {
		stages[name] = system.NewStage(name)
	}
	server := &ChannelServer{
		ID:              config.ID,
		logger:          logger.Get().Named("channel-" + fmt.Sprint(config.ID)),
		acceptConns:     make(chan net.Conn),
		deleteConns:     make(chan net.Conn),
		sessions:        make(map[net.Conn]*Session),
		objectIDs:       make(map[*Session]uint16),
		stages:          stages,
		userBinaryParts: make(map[userBinaryPartID][]byte),
		semaphore:       make(map[string]*Semaphore),
		discordBot:      config.DiscordBot,
		name:            config.Name,
		raviente: &Raviente{
			id:       1,
			register: make([]uint32, 30),
			state:    make([]uint32, 30),
			support:  make([]uint32, 30),
		},
		questCacheData: make(map[int][]byte),
		questCacheTime: make(map[int]time.Time),
	}
	server.initCommands()

	server.i18n = getLangStrings(server)
	return server
}

// Start starts the server in a new goroutine.
func (server *ChannelServer) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", server.Port))
	if err != nil {
		return err
	}
	server.listener = l

	go server.acceptClients()
	go server.manageSessions()
	// Start the discord bot for chat integration.
	if config.GetConfig().Discord.Enabled && server.discordBot != nil {
		server.discordBot.Session.AddHandler(server.onDiscordMessage)
		server.discordBot.Session.AddHandler(server.onInteraction)
	}

	return nil
}

// Shutdown tries to shut down the server gracefully.
func (server *ChannelServer) Shutdown() {
	server.Lock()
	server.isShuttingDown = true
	server.Unlock()

	server.listener.Close()

	close(server.acceptConns)
}

func (server *ChannelServer) acceptClients() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			server.Lock()
			shutdown := server.isShuttingDown
			server.Unlock()

			if shutdown {
				break
			} else {
				server.logger.Warn("Error accepting client", zap.Error(err))
				continue
			}
		}
		server.acceptConns <- conn
	}
}

func (server *ChannelServer) manageSessions() {
	for {
		select {
		case newConn := <-server.acceptConns:
			// Gracefully handle acceptConns channel closing.
			if newConn == nil {
				server.Lock()
				shutdown := server.isShuttingDown
				server.Unlock()

				if shutdown {
					return
				}
			}

			session := NewSession(server, newConn)

			server.Lock()
			server.sessions[newConn] = session
			server.Unlock()

			session.Start()

		case delConn := <-server.deleteConns:
			server.Lock()
			delete(server.sessions, delConn)
			server.Unlock()
		}
	}
}

func (server *ChannelServer) FindSessionByCharID(charID uint32) *Session {
	for _, c := range server.Channels {
		for _, session := range c.sessions {
			if session.CharID == charID {
				return session
			}
		}
	}
	return nil
}

func (server *ChannelServer) DisconnectUser(uid uint32) {
	var cid uint32
	var cids []uint32
	db, err := db.GetDB()
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	rows, _ := db.Query(`SELECT id FROM characters WHERE user_id=$1`, uid)
	for rows.Next() {
		rows.Scan(&cid)
		cids = append(cids, cid)
	}
	for _, c := range server.Channels {
		for _, session := range c.sessions {
			for _, cid := range cids {
				if session.CharID == cid {
					session.rawConn.Close()
					break
				}
			}
		}
	}
}

func (server *ChannelServer) FindObjectByChar(charID uint32) *system.Object {
	server.stagesLock.RLock()
	defer server.stagesLock.RUnlock()
	for _, stage := range server.stages {
		stage.RLock()
		for objId := range stage.Objects {
			obj := stage.Objects[objId]
			if obj.OwnerCharID == charID {
				stage.RUnlock()
				return obj
			}
		}
		stage.RUnlock()
	}

	return nil
}

func (server *ChannelServer) HasSemaphore(ses *Session) bool {
	for _, semaphore := range server.semaphore {
		if semaphore.host == ses {
			return true
		}
	}
	return false
}

func (server *ChannelServer) Season() uint8 {
	sid := int64(((server.ID & 0xFF00) - 4096) / 256)
	return uint8(((gametime.TimeAdjusted().Unix() / 86400) + sid) % 3)
}
