package database

import (
	"fmt"
	"sync"

	"erupe-ce/config"
	"erupe-ce/utils/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver
)

var (
	instance *sqlx.DB
	once     sync.Once
	dbLogger logger.Logger
)

// InitDB initializes the database connection pool as a singleton
func InitDB(config *config.Config) (*sqlx.DB, error) {
	dbLogger = logger.Get().Named("database")
	var err error
	once.Do(func() {
		// Create the postgres DB pool.
		connectString := fmt.Sprintf(
			"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
			config.Database.Host,
			config.Database.Port,
			config.Database.User,
			config.Database.Password,
			config.Database.Database,
		)

		instance, err = sqlx.Open("postgres", connectString)
		if err != nil {
			return // Stop here if there's an error opening the database
		}

		// Test the DB connection.
		err = instance.Ping()
		if err != nil {
			return // Stop here if there's an error pinging the database
		}
		dbLogger.Info("Database: Started successfully")

		// Clear stale data
		if config.DebugOptions.ProxyPort == 0 {
			_ = instance.MustExec("DELETE FROM sign_sessions")
		}
		_ = instance.MustExec("DELETE FROM servers")
		_ = instance.MustExec(`UPDATE guild_characters SET treasure_hunt=NULL`)

		// Clean the DB if the option is on.
		if config.DebugOptions.CleanDB {
			dbLogger.Info("Database: Started clearing...")
			cleanDB(instance)
			dbLogger.Info("Database: Finished clearing")
		}
	})
	return instance, err
}

func GetDB() (*sqlx.DB, error) {
	if instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return instance, nil
}

// Temporary DB auto clean on startup for quick development & testing.
func cleanDB(db *sqlx.DB) {
	_ = db.MustExec("DELETE FROM guild_characters")
	_ = db.MustExec("DELETE FROM guilds")
	_ = db.MustExec("DELETE FROM characters")
	_ = db.MustExec("DELETE FROM users")
}
