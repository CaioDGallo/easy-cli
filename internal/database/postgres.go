package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/interfaces"
	"github.com/CaioDGallo/easy-cli/internal/logger"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var _ interfaces.DatabaseProvider = (*PostgresService)(nil)

type PostgresService struct {
	config config.DatabaseConfig
}

func NewPostgresService(cfg config.DatabaseConfig) *PostgresService {
	return &PostgresService{
		config: cfg,
	}
}

func (p *PostgresService) CreateClientDatabase(mainDBName string) error {
	return p.CreateClientDatabases(mainDBName, fmt.Sprintf("%s-hf", mainDBName))
}

func (p *PostgresService) CreateClientDatabases(mainDBName, hangfireDBName string) error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := p.killConnections(db, "demo"); err != nil {
		return fmt.Errorf("failed to kill connections to template database: %w", err)
	}

	if err := p.createDatabase(db, mainDBName, "demo"); err != nil {
		return fmt.Errorf("failed to create main database: %w", err)
	}

	if err := p.killConnections(db, "demo-hf"); err != nil {
		return fmt.Errorf("failed to kill connections to hangfire template database: %w", err)
	}

	if err := p.createDatabase(db, hangfireDBName, "demo-hf"); err != nil {
		return fmt.Errorf("failed to create hangfire database: %w", err)
	}

	return nil
}

func (p *PostgresService) killConnections(db *sql.DB, templateDB string) error {
	query := `SELECT pg_terminate_backend(pg_stat_activity.pid) 
			  FROM pg_stat_activity 
			  WHERE pg_stat_activity.datname = $1 AND pid <> pg_backend_pid()`

	_, err := db.Exec(query, templateDB)
	return err
}

func (p *PostgresService) createDatabase(db *sql.DB, dbName, templateDB string) error {
	query := fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE "%s"`, dbName, templateDB)
	_, err := db.Exec(query)
	return err
}

func (p *PostgresService) DeleteClientDatabases(mainDBName, hangfireDBName string) error {
	log := logger.WithFields(logrus.Fields{
		"main_db":     mainDBName,
		"hangfire_db": hangfireDBName,
		"service":     "postgres",
		"action":      "delete",
	})

	log.Info("Starting database deletion")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.WithError(err).Error("Failed to connect to database")
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.WithError(err).Error("Failed to ping database")
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := p.deleteDatabase(db, mainDBName); err != nil {
		if !p.isDatabaseNotFoundError(err) {
			log.WithError(err).Error("Failed to delete main database")
			return fmt.Errorf("failed to delete main database %s: %w", mainDBName, err)
		}
		log.Info("Main database does not exist, skipping deletion")
	} else {
		log.Info("Main database deleted successfully")
	}

	if err := p.deleteDatabase(db, hangfireDBName); err != nil {
		if !p.isDatabaseNotFoundError(err) {
			log.WithError(err).Error("Failed to delete hangfire database")
			return fmt.Errorf("failed to delete hangfire database %s: %w", hangfireDBName, err)
		}
		log.Info("Hangfire database does not exist, skipping deletion")
	} else {
		log.Info("Hangfire database deleted successfully")
	}

	log.Info("Database deletion completed successfully")
	return nil
}

func (p *PostgresService) deleteDatabase(db *sql.DB, dbName string) error {
	log := logger.WithFields(logrus.Fields{
		"database": dbName,
		"service":  "postgres",
	})

	log.Info("Terminating connections to database")

	if err := p.killConnections(db, dbName); err != nil {
		log.WithError(err).Warn("Failed to kill connections, continuing with deletion")
	}

	log.Info("Dropping database")

	query := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, dbName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	return nil
}

func (p *PostgresService) isDatabaseNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "database not found")
}
