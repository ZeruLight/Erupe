package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"math/rand"

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

// Temporary DB auto clean on startup for quick development & testing.
func cleanDB(db *sqlx.DB) {
	_ = db.MustExec("DELETE FROM guild_characters")
	_ = db.MustExec("DELETE FROM guilds")
	_ = db.MustExec("DELETE FROM characters")
	_ = db.MustExec("DELETE FROM sign_sessions")
	_ = db.MustExec("DELETE FROM users")
}

func main() {
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger := zapLogger.Named("main")

	logger.Info("Starting Erupe")

	// Load the configuration.
	erupeConfig, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Discord bot
	var discordBot *discordbot.DiscordBot = nil

	if erupeConfig.Discord.Enabled {
		bot, err := discordbot.NewDiscordBot(discordbot.DiscordBotOptions{
			Logger: logger,
			Config: erupeConfig,
		})

		if err != nil {
			logger.Fatal("Failed to create discord bot", zap.Error(err))
		}

		// Discord bot
		err = bot.Start()

		if err != nil {
			logger.Fatal("Failed to starts discord bot", zap.Error(err))
		}

		discordBot = bot
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
		logger.Fatal("Failed to open sql database", zap.Error(err))
	}

	// Test the DB connection.
	err = db.Ping()
	if err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// Clear existing tokens
	_ = db.MustExec("DELETE FROM sign_sessions")
	_ = db.MustExec("DELETE FROM servers")

	// Clean the DB if the option is on.
	if erupeConfig.DevMode && erupeConfig.DevModeOptions.CleanDB {
		logger.Info("Cleaning DB")
		cleanDB(db)
		logger.Info("Done cleaning DB")
	}

	// Now start our server(s).

	// Launcher HTTP server.
	launcherServer := launcherserver.NewServer(
		&launcherserver.Config{
			Logger:                   logger.Named("launcher"),
			ErupeConfig:              erupeConfig,
			DB:                       db,
			UseOriginalLauncherFiles: erupeConfig.Launcher.UseOriginalLauncherFiles,
		})
	err = launcherServer.Start()
	if err != nil {
		logger.Fatal("Failed to start launcher server", zap.Error(err))
	}
	logger.Info("Started launcher server")

	// Entrance server.
	entranceServer := entranceserver.NewServer(
		&entranceserver.Config{
			Logger:      logger.Named("entrance"),
			ErupeConfig: erupeConfig,
			DB:          db,
		})
	err = entranceServer.Start()
	if err != nil {
		logger.Fatal("Failed to start entrance server", zap.Error(err))
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
		logger.Fatal("Failed to start sign server", zap.Error(err))
	}
	logger.Info("Started sign server")

	var channels []*channelserver.Server
	channelQuery := ""
	si := 0
	ci := 0
	count := 1
	for _, ee := range erupeConfig.Entrance.Entries {
		rand.Seed(time.Now().UnixNano())
		// Randomly generate a season for the World
		season := rand.Intn(3)+1
		for _, ce := range ee.Channels {
			sid := (4096 + si * 256) + (16 + ci)
			c := *channelserver.NewServer(&channelserver.Config{
				ID:           uint16(sid),
				Logger:       logger.Named("channel-"+fmt.Sprint(count)),
				ErupeConfig:  erupeConfig,
				DB:           db,
				DiscordBot:   discordBot,
			})
			err = c.Start(int(ce.Port))
			if err != nil {
				logger.Fatal("Failed to start channel", zap.Error(err))
			} else {
				channelQuery += fmt.Sprintf("INSERT INTO servers (server_id, season, current_players) VALUES (%d, %d, 0);", sid, season)
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
	launcherServer.Shutdown()

	time.Sleep(1 * time.Second)
}
