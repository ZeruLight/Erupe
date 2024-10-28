package main

import (
	"erupe-ce/config"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"erupe-ce/server/api"
	"erupe-ce/server/channelserver"
	"erupe-ce/server/discordbot"
	"erupe-ce/server/entrance"
	"erupe-ce/server/sign"
	"erupe-ce/utils/database"
	"erupe-ce/utils/logger"

	"erupe-ce/utils/gametime"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var mainLogger logger.Logger

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:7]
			}
		}
	}
	return "unknown"
}

func initLogger() {
	var zapLogger *zap.Logger
	zapLogger, _ = zap.NewDevelopment(zap.WithCaller(false))
	defer zapLogger.Sync()
	// Initialize the global logger
	logger.Init(zapLogger)
	mainLogger = logger.Get().Named("main")

}

func main() {
	var err error

	config := config.GetConfig()
	initLogger()
	mainLogger.Info(fmt.Sprintf("Starting Erupe (9.3b-%s)", Commit()))
	mainLogger.Info(fmt.Sprintf("Client Mode: %s (%d)", config.ClientMode, config.ClientID))

	checkAndExitIf(config.Database.Password == "", "Database password is blank")

	resolveHostIP()

	discordBot := initializeDiscordBot()

	database, err := database.InitDB(config)
	if err != nil {
		mainLogger.Fatal(fmt.Sprintf("Database initialization failed: %s", err))
	}

	mainLogger.Info(fmt.Sprintf("Server Time: %s", gametime.TimeAdjusted().String()))

	// Now start our server(s).

	// Entrance server.

	var entranceServer *entrance.EntranceServer
	if config.Entrance.Enabled {
		entranceServer = entrance.NewServer()
		err = entranceServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Entrance: Failed to start, %s", err.Error()))
		}
		mainLogger.Info("Entrance: Started successfully")
	} else {
		mainLogger.Info("Entrance: Disabled")
	}

	// Sign server.
	var signServer *sign.SignServer
	if config.Sign.Enabled {
		signServer = sign.NewServer()
		err = signServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Sign: Failed to start, %s", err.Error()))
		}
		mainLogger.Info("Sign: Started successfully")
	} else {
		mainLogger.Info("Sign: Disabled")
	}

	// Api server
	var apiServer *api.APIServer
	if config.API.Enabled {
		apiServer = api.NewAPIServer()
		err = apiServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("API: Failed to start, %s", err.Error()))
		}
		mainLogger.Info("API: Started successfully")
	} else {
		mainLogger.Info("API: Disabled")
	}
	var channelServers []*channelserver.ChannelServer
	if config.Channel.Enabled {
		channelQuery := ""
		si := 0
		ci := 0
		count := 1
		for j, ee := range config.Entrance.Entries {
			for i, ce := range ee.Channels {
				sid := (4096 + si*256) + (16 + ci)
				c := *channelserver.NewServer(&channelserver.Config{
					ID:         uint16(sid),
					DiscordBot: discordBot,
				})
				if ee.IP == "" {
					c.IP = config.Host
				} else {
					c.IP = ee.IP
				}
				c.Port = ce.Port
				c.GlobalID = fmt.Sprintf("%02d%02d", j+1, i+1)
				err = c.Start()
				if err != nil {
					preventClose(fmt.Sprintf("Channel: Failed to start, %s", err.Error()))
				} else {
					channelQuery += fmt.Sprintf(`INSERT INTO servers (server_id, current_players, world_name, world_description, land) VALUES (%d, 0, '%s', '%s', %d);`, sid, ee.Name, ee.Description, i+1)
					channelServers = append(channelServers, &c)
					mainLogger.Info(fmt.Sprintf("Channel %d (%d): Started successfully", count, ce.Port))
					ci++
					count++
				}
			}
			ci = 0
			si++
		}

		// Register all servers in DB
		_ = database.MustExec(channelQuery)

		for _, c := range channelServers {
			c.Channels = channelServers
		}
	}

	mainLogger.Info("Finished starting Erupe")

	// Wait for exit or interrupt with ctrl+C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	if !config.DisableSoftCrash {
		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("Shutting down in %d...", 10-i)
			for _, channelServer := range channelServers {
				channelServer.BroadcastChatMessage(message)
			}
			mainLogger.Warn(message)
			time.Sleep(time.Second)
		}
	}

	if config.Channel.Enabled {
		for _, channelServer := range channelServers {
			channelServer.Shutdown()
		}
	}

	if config.Sign.Enabled {
		signServer.Shutdown()
	}

	if config.API.Enabled {
		apiServer.Shutdown()
	}

	if config.Entrance.Enabled {
		entranceServer.Shutdown()
	}

	time.Sleep(1 * time.Second)
}

func wait() {
	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func preventClose(text string) {
	if config.GetConfig().DisableSoftCrash {
		os.Exit(0)
	}
	mainLogger.Error(fmt.Sprintf(("\nFailed to start Erupe:\n" + text)))
	go wait()
	mainLogger.Error(fmt.Sprintf(("\nPress Enter/Return to exit...")))
	os.Exit(0)
}

func checkAndExitIf(condition bool, message string) {
	if condition {
		preventClose(message)
	}
}

func resolveHostIP() {
	if net.ParseIP(config.GetConfig().Host) == nil {
		ips, err := net.LookupIP(config.GetConfig().Host)
		if err == nil && len(ips) > 0 {
			config.GetConfig().Host = ips[0].String()
		}
		checkAndExitIf(net.ParseIP(config.GetConfig().Host) == nil, "Invalid host address")
	}
}

func initializeDiscordBot() *discordbot.DiscordBot {
	if !config.GetConfig().Discord.Enabled {
		mainLogger.Info("Discord: Disabled")
		return nil
	}

	bot, err := discordbot.NewDiscordBot()
	checkAndExitIf(err != nil, fmt.Sprintf("Discord: Failed to start, %s", err))

	err = bot.Start()
	checkAndExitIf(err != nil, fmt.Sprintf("Discord: Failed to start, %s", err))

	_, err = bot.Session.ApplicationCommandBulkOverwrite(bot.Session.State.User.ID, "", discordbot.Commands)
	checkAndExitIf(err != nil, fmt.Sprintf("Discord: Failed to register commands, %s", err))

	mainLogger.Info("Discord: Started successfully")
	return bot
}
