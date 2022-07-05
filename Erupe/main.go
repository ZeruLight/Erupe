package main

import (
	"fmt"
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
	logger.Info("Started launcher server.")

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
	logger.Info("Started entrance server.")

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
	logger.Info("Started sign server.")

	// Channel Server
	channelServer1 := channelserver.NewServer(
		&channelserver.Config{
			Logger:      logger.Named("channel"),
			ErupeConfig: erupeConfig,
			DB:          db,
			Name:        erupeConfig.Entrance.Entries[0].Name,
			Enable:      erupeConfig.Entrance.Entries[0].Channels[0].MaxPlayers > 0,
			//DiscordBot:  discordBot,
		})

	err = channelServer1.Start(erupeConfig.Channel.Port1)
	if err != nil {
		logger.Fatal("Failed to start channel server1", zap.Error(err))
	}
	logger.Info("Started channel server.")
	// Channel Server
	channelServer2 := channelserver.NewServer(
		&channelserver.Config{
			Logger:      logger.Named("channel"),
			ErupeConfig: erupeConfig,
			DB:          db,
			Name:        erupeConfig.Entrance.Entries[1].Name,
			Enable:      erupeConfig.Entrance.Entries[1].Channels[0].MaxPlayers > 0,
			DiscordBot:  discordBot,
		})

	err = channelServer2.Start(erupeConfig.Channel.Port2)
	if err != nil {
		logger.Fatal("Failed to start channel server2", zap.Error(err))
	}
	// Channel Server
	channelServer3 := channelserver.NewServer(
		&channelserver.Config{
			Logger:      logger.Named("channel"),
			ErupeConfig: erupeConfig,
			DB:          db,
			Name:        erupeConfig.Entrance.Entries[2].Name,
			Enable:      erupeConfig.Entrance.Entries[2].Channels[0].MaxPlayers > 0,
			//DiscordBot:  discordBot,
		})

	err = channelServer3.Start(erupeConfig.Channel.Port3)
	if err != nil {
		logger.Fatal("Failed to start channel server3", zap.Error(err))
	}
	// Channel Server
	channelServer4 := channelserver.NewServer(
		&channelserver.Config{
			Logger:      logger.Named("channel"),
			ErupeConfig: erupeConfig,
			DB:          db,
			Name:        erupeConfig.Entrance.Entries[3].Name,
			Enable:      erupeConfig.Entrance.Entries[3].Channels[0].MaxPlayers > 0,
			//DiscordBot:  discordBot,
		})

	err = channelServer4.Start(erupeConfig.Channel.Port4)
	if err != nil {
		logger.Fatal("Failed to start channel server4", zap.Error(err))
	}
	// Wait for exit or interrupt with ctrl+C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Trying to shutdown gracefully.")
	channelServer4.Shutdown()
	channelServer3.Shutdown()
	channelServer2.Shutdown()
	channelServer1.Shutdown()
	signServer.Shutdown()
	entranceServer.Shutdown()
	launcherServer.Shutdown()

	time.Sleep(1 * time.Second)
}
