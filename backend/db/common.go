package db

import (
	"fmt"
	"sync"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/syjn99/leanView/backend/types"
)

// Global db instances
var ReaderDb *sqlx.DB
var writerDb *sqlx.DB
var writerMutex sync.Mutex

var logger = logrus.StandardLogger().WithField("module", "db")

func InitDB(cfg *types.DatabaseConfig) {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.MaxOpenConns < cfg.MaxIdleConns {
		cfg.MaxIdleConns = cfg.MaxOpenConns
	}

	logger.Infof("Initializing sqlite connection to %v with %v/%v conn limit", cfg.File, cfg.MaxIdleConns, cfg.MaxOpenConns)
	dbConn, err := sqlx.Open("sqlite", fmt.Sprintf("%s?_pragma=journal_mode(WAL)", cfg.File))
	if err != nil {
		logger.WithError(err).Fatal("error opening sqlite database")
	}

	checkDbConn(dbConn, "database")
	dbConn.SetConnMaxIdleTime(0)
	dbConn.SetConnMaxLifetime(0)
	dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
	dbConn.SetMaxIdleConns(cfg.MaxIdleConns)

	dbConn.MustExec("PRAGMA journal_mode = WAL")

	ReaderDb = dbConn
	writerDb = dbConn

	// Run database migrations
	if err := RunMigrations(); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}
}

func checkDbConn(dbConn *sqlx.DB, dataBaseName string) {
	// The golang sql driver does not properly implement PingContext
	// therefore we use a timer to catch db connection timeouts
	dbConnectionTimeout := time.NewTimer(15 * time.Second)

	go func() {
		<-dbConnectionTimeout.C
		logger.Fatalf("timeout while connecting to %s", dataBaseName)
	}()

	err := dbConn.Ping()
	if err != nil {
		logger.Fatalf("unable to Ping %s: %s", dataBaseName, err)
	}

	dbConnectionTimeout.Stop()
}

func RunDBTransaction(handler func(tx *sqlx.Tx) error) error {
	writerMutex.Lock()
	defer writerMutex.Unlock()

	tx, err := writerDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}

	defer tx.Rollback()

	err = handler(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %v", err)
	}

	return nil
}
