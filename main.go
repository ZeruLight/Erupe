package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"erupe-ce/config"
	"erupe-ce/server/channelserver"
	"erupe-ce/server/discordbot"
	"erupe-ce/server/entranceserver"
	"erupe-ce/server/launcherserver"
	"erupe-ce/server/signserver"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var erupeConfig *config.Config

// Temporary DB auto clean on startup for quick development & testing.
func cleanDB(db *sqlx.DB) {
	_ = db.MustExec("DELETE FROM guild_characters")
	_ = db.MustExec("DELETE FROM guilds")
	_ = db.MustExec("DELETE FROM characters")
	_ = db.MustExec("DELETE FROM sign_sessions")
	_ = db.MustExec("DELETE FROM users")
}

func main() {
	var err error
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger := zapLogger.Named("main")

	logger.Info("Starting Erupe")

	// Load the configuration.
	erupeConfig, err = config.LoadConfig()
	if err != nil {
		preventClose(fmt.Sprintf("Failed to load config: %s", err.Error()))
	}

	if erupeConfig.Database.Password == "" {
		preventClose("Database password is blank")
	}

	if net.ParseIP(erupeConfig.Host) == nil {
		ips, _ := net.LookupIP(erupeConfig.Host)
		for _, ip := range ips {
			if ip != nil {
				erupeConfig.Host = ip.String()
				break
			}
		}
		if net.ParseIP(erupeConfig.Host) == nil {
			preventClose("Invalid host address")
		}
	}

	// Discord bot
	var discordBot *discordbot.DiscordBot = nil

	if erupeConfig.Discord.Enabled {
		bot, err := discordbot.NewDiscordBot(discordbot.Options{
			Logger: logger,
			Config: erupeConfig,
		})

		if err != nil {
			preventClose(fmt.Sprintf("Failed to create Discord bot: %s", err.Error()))
		}

		// Discord bot
		err = bot.Start()

		if err != nil {
			preventClose(fmt.Sprintf("Failed to start Discord bot: %s", err.Error()))
		}

		discordBot = bot
		logger.Info("Discord bot is enabled")
	} else {
		logger.Info("Discord bot is disabled")
	}

	// Create the postgres DB pool.
	connectString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname= %s sslmode=disable",
		erupeConfig.Database.Host,
		erupeConfig.Database.Port,
		erupeConfig.Database.User,
		erupeConfig.Database.Password,
		erupeConfig.Database.Database,
	)

	db, err := sqlx.Open("postgres", connectString)
	if err != nil {
		preventClose(fmt.Sprintf("Failed to open SQL database: %s", err.Error()))
	}

	// Test the DB connection.
	err = db.Ping()
	if err != nil {
		preventClose(fmt.Sprintf("Failed to ping database: %s", err.Error()))
	}
	logger.Info("Connected to database")

	// Clear stale data
	_ = db.MustExec("DELETE FROM sign_sessions")
	_ = db.MustExec("DELETE FROM servers")
	_ = db.MustExec("DELETE FROM cafe_accepted")
	_ = db.MustExec("UPDATE characters SET cafe_time=0")

	// Clean the DB if the option is on.
	if erupeConfig.DevMode && erupeConfig.DevModeOptions.CleanDB {
		logger.Info("Cleaning DB")
		cleanDB(db)
		logger.Info("Done cleaning DB")
	}

	// Now start our server(s).

	// Launcher HTTP server.
	var launcherServer *launcherserver.Server
	if erupeConfig.DevMode && erupeConfig.DevModeOptions.EnableLauncherServer {
		launcherServer = launcherserver.NewServer(
			&launcherserver.Config{
				Logger:                   logger.Named("launcher"),
				ErupeConfig:              erupeConfig,
				DB:                       db,
				UseOriginalLauncherFiles: erupeConfig.Launcher.UseOriginalLauncherFiles,
			})
		err = launcherServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Failed to start launcher server: %s", err.Error()))
		}
		logger.Info("Started launcher server")
	}

	// Entrance server.
	entranceServer := entranceserver.NewServer(
		&entranceserver.Config{
			Logger:      logger.Named("entrance"),
			ErupeConfig: erupeConfig,
			DB:          db,
		})
	err = entranceServer.Start()
	if err != nil {
		preventClose(fmt.Sprintf("Failed to start entrance server: %s", err.Error()))
	}
	logger.Info("Started entrance server")

	// Sign server.
	signServer := signserver.NewServer(
		&signserver.Config{
			Logger:      logger.Named("sign"),
			ErupeConfig: erupeConfig,
			DB:          db,
		})
	err = signServer.Start()
	if err != nil {
		preventClose(fmt.Sprintf("Failed to start sign server: %s", err.Error()))
	}
	logger.Info("Started sign server")

	var channels []*channelserver.Server
	channelQuery := ""
	si := 0
	ci := 0
	count := 1
	for _, ee := range erupeConfig.Entrance.Entries {
		for _, ce := range ee.Channels {
			sid := (4096 + si*256) + (16 + ci)
			c := *channelserver.NewServer(&channelserver.Config{
				ID:          uint16(sid),
				Logger:      logger.Named("channel-" + fmt.Sprint(count)),
				ErupeConfig: erupeConfig,
				DB:          db,
				DiscordBot:  discordBot,
			})
			if ee.IP == "" {
				c.IP = erupeConfig.Host
			} else {
				c.IP = ee.IP
			}
			c.Port = ce.Port
			err = c.Start()
			if err != nil {
				preventClose(fmt.Sprintf("Failed to start channel server: %s", err.Error()))
			} else {
				channelQuery += fmt.Sprintf("INSERT INTO servers (server_id, season, current_players) VALUES (%d, %d, 0);", sid, si%3)
				channels = append(channels, &c)
				logger.Info(fmt.Sprintf("Started channel server %d on port %d", count, ce.Port))
				ci++
				count++
			}
		}
		ci = 0
		si++
	}

	// Register all servers in DB
	_ = db.MustExec(channelQuery)

	for _, c := range channels {
		c.Channels = channels
	}

	// Wait for exit or interrupt with ctrl+C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Trying to shutdown gracefully")

	for _, c := range channels {
		c.Shutdown()
	}
	signServer.Shutdown()
	entranceServer.Shutdown()
	if erupeConfig.DevModeOptions.EnableLauncherServer {
		launcherServer.Shutdown()
	}

	time.Sleep(1 * time.Second)
}

func wait() {
	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func preventClose(text string) {
	if erupeConfig.DisableSoftCrash {
		os.Exit(0)
	}
	fmt.Println("\nFailed to start Erupe:\n" + text)
	go wait()
	fmt.Println("\nPress Enter/Return to exit...")
	fmt.Scanln()
	os.Exit(0)
}
