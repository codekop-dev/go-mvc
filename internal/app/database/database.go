package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codekop/gomvc/internal/app/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open creates and pings a database/sql pool for MySQL/MariaDB or PostgreSQL.
func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	// A blank DB_NAME keeps the stateless sample API runnable before a database exists.
	if cfg.Name == "" {
		return nil, nil
	}

	dsn, err := dsn(cfg)
	if err != nil {
		return nil, err
	}
	driverName := cfg.Driver
	if driverName == "postgres" {
		driverName = "pgx"
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cfg.Driver, err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping %s: %w", cfg.Driver, err)
	}
	return db, nil
}

// dsn builds the driver-specific connection string from application config.
func dsn(cfg config.DatabaseConfig) (string, error) {
	switch cfg.Driver {
	case "mysql":
		if cfg.Socket != "" {
			return fmt.Sprintf("%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.User, cfg.Password, cfg.Socket, cfg.Name), nil
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name), nil
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode), nil
	default:
		return "", fmt.Errorf("unsupported DB_DRIVER %q; use mysql or postgres", cfg.Driver)
	}
}
