package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/jp-roisin/catch-and-go/internal/database/seeds"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dburl := os.Getenv("BLUEPRINT_DB_URL")
	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		log.Fatalf("❌ Failed to open DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err := seeds.SeedStops(ctx, db); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Seeding complete")
}
