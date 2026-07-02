package database

import (
	"fmt"
	"log"
	"time"

	"github.com/pmfb/sige-ti/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect abre a conexão com o banco de dados escolhendo o dialeto do GORM
// conforme a configuração. A troca SQLite <-> PostgreSQL é apenas questão
// de configuração (DB_DRIVER e DB_DSN), sem alterar código.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dialector, err := buildDialector(cfg)
	if err != nil {
		return nil, err
	}

	gormCfg := &gorm.Config{
		Logger: gormLogger(cfg.GinMode),
		NowFunc: func() time.Time {
			// Mantém timestamps em UTC para consistência entre ambientes.
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco (%s): %w", cfg.DBDriver, err)
	}

	if err := tuneConnectionPool(db, cfg); err != nil {
		return nil, err
	}

	log.Printf("[database] conectado via dialeto %q", cfg.DBDriver)
	return db, nil
}

// buildDialector seleciona o driver do GORM de acordo com DB_DRIVER.
func buildDialector(cfg *config.Config) (gorm.Dialector, error) {
	switch cfg.DBDriver {
	case "sqlite":
		return sqlite.Open(cfg.DBDSN), nil
	case "postgres":
		return postgres.Open(cfg.DBDSN), nil
	default:
		return nil, fmt.Errorf("driver de banco não suportado: %q", cfg.DBDriver)
	}
}

func tuneConnectionPool(db *gorm.DB, cfg *config.Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("falha ao obter *sql.DB: %w", err)
	}

	if cfg.DBDriver == "sqlite" {
		// SQLite com um único writer: limitamos a 1 conexão de escrita para
		// evitar "database is locked". Habilitamos WAL para melhor concorrência.
		sqlDB.SetMaxOpenConns(1)
		if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
			return fmt.Errorf("falha ao habilitar WAL: %w", err)
		}
		if err := db.Exec("PRAGMA foreign_keys=ON;").Error; err != nil {
			return fmt.Errorf("falha ao habilitar foreign_keys: %w", err)
		}
	} else {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	return nil
}

func gormLogger(ginMode string) logger.Interface {
	if ginMode == "release" {
		return logger.Default.LogMode(logger.Warn)
	}
	return logger.Default.LogMode(logger.Info)
}
