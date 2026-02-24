package postgres

import (
	"fmt"
	"log"
	"time"

	"coffee-shop-pos/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewConnection(cfg *configs.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to PostgreSQL successfully")
	return db, nil
}
