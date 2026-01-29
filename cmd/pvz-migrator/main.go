package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"pvz-service/internal/config"
)

func main() {
	cfg := config.Load()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB: ", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatal("Failed to apply migrations: ", err)
	}
	log.Println("Migrations applied successfully")
}
