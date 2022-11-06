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
	"erupe-ce/server/signv2server"

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
	var err error
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger := zapLogger.Named("main")

	logger.Info("Starting Erupe (9.1)")

	if config.ErupeConfig.Database.Password == "" {
		preventClose("Database password is blank")
	}

	if net.ParseIP(config.ErupeConfig.Host) == nil {
		ips, _ := net.LookupIP(config.ErupeConfig.Host)
		for _, ip := range ips {
			if ip != nil {
				config.ErupeConfig.Host = ip.String()
				break
			}
		}
		if net.ParseIP(config.ErupeConfig.Host) == nil {
			preventClose("Invalid host address")
		}
	}

	// Discord bot
	var discordBot *discordbot.DiscordBot = nil

	if config.ErupeConfig.Discord.Enabled {
		bot, err := discordbot.NewDiscordBot(discordbot.Options{
			Logger: logger,
			Config: config.ErupeConfig,
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
		config.ErupeConfig.Database.Host,
		config.ErupeConfig.Database.Port,
		config.ErupeConfig.Database.User,
		config.ErupeConfig.Database.Password,
		config.ErupeConfig.Database.Database,
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

	// Clean the DB if the option is on.
	if config.ErupeConfig.DevMode && config.ErupeConfig.DevModeOptions.CleanDB {
		logger.Info("Cleaning DB")
		cleanDB(db)
		logger.Info("Done cleaning DB")
	}

	// Now start our server(s).

	// Launcher HTTP server.
	var launcherServer *launcherserver.Server
	if config.ErupeConfig.Launcher.Enabled {
		launcherServer = launcherserver.NewServer(
			&launcherserver.Config{
				Logger:                   logger.Named("launcher"),
				ErupeConfig:              config.ErupeConfig,
				DB:                       db,
				UseOriginalLauncherFiles: config.ErupeConfig.Launcher.UseOriginalLauncherFiles,
			})
		err = launcherServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Failed to start launcher server: %s", err.Error()))
		}
		logger.Info("Started launcher server")
	}

	// Entrance server.

	var entranceServer *entranceserver.Server
	if config.ErupeConfig.Entrance.Enabled {
		entranceServer = entranceserver.NewServer(
			&entranceserver.Config{
				Logger:      logger.Named("entrance"),
				ErupeConfig: config.ErupeConfig,
				DB:          db,
			})
		err = entranceServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Failed to start entrance server: %s", err.Error()))
		}
		logger.Info("Started entrance server")
	}

	// Sign server.

	var signServer *signserver.Server
	if config.ErupeConfig.Sign.Enabled {
		signServer = signserver.NewServer(
			&signserver.Config{
				Logger:      logger.Named("sign"),
				ErupeConfig: config.ErupeConfig,
				DB:          db,
			})
		err = signServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Failed to start sign server: %s", err.Error()))
		}
		logger.Info("Started sign server")
	}

	// New Sign server
	var newSignServer *signv2server.Server
	if config.ErupeConfig.SignV2.Enabled {
		newSignServer = signv2server.NewServer(
			&signv2server.Config{
				Logger:      logger.Named("sign"),
				ErupeConfig: config.ErupeConfig,
				DB:          db,
			})
		err = newSignServer.Start()
		if err != nil {
			preventClose(fmt.Sprintf("Failed to start sign-v2 server: %s", err.Error()))
		}
		logger.Info("Started new sign server")
	}

	var channels []*channelserver.Server

	if config.ErupeConfig.Channel.Enabled {
		channelQuery := ""
		si := 0
		ci := 0
		count := 1
		for _, ee := range config.ErupeConfig.Entrance.Entries {
			for i, ce := range ee.Channels {
				sid := (4096 + si*256) + (16 + ci)
				c := *channelserver.NewServer(&channelserver.Config{
					ID:          uint16(sid),
					Logger:      logger.Named("channel-" + fmt.Sprint(count)),
					ErupeConfig: config.ErupeConfig,
					DB:          db,
					DiscordBot:  discordBot,
				})
				if ee.IP == "" {
					c.IP = config.ErupeConfig.Host
				} else {
					c.IP = ee.IP
				}
				c.Port = ce.Port
				err = c.Start()
				if err != nil {
					preventClose(fmt.Sprintf("Failed to start channel server: %s", err.Error()))
				} else {
					channelQuery += fmt.Sprintf(`INSERT INTO servers (server_id, season, current_players, world_name, world_description, land) VALUES (%d, %d, 0, '%s', '%s', %d);`, sid, si%3, ee.Name, ee.Description, i+1)
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
	}

	// Wait for exit or interrupt with ctrl+C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Trying to shutdown gracefully")

	if config.ErupeConfig.Channel.Enabled {
		for _, c := range channels {
			c.Shutdown()
		}
	}

	if config.ErupeConfig.Sign.Enabled {
		signServer.Shutdown()
	}

	if config.ErupeConfig.SignV2.Enabled {
		newSignServer.Shutdown()
	}

	if config.ErupeConfig.Entrance.Enabled {
		entranceServer.Shutdown()
	}

	if config.ErupeConfig.Launcher.Enabled {
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
	if config.ErupeConfig.DisableSoftCrash {
		os.Exit(0)
	}
	fmt.Println("\nFailed to start Erupe:\n" + text)
	go wait()
	fmt.Println("\nPress Enter/Return to exit...")
	fmt.Scanln()
	os.Exit(0)
}
