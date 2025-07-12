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
		log.Fatalf("❌ Stops seeding failed:\n %v", err)
	}

	if err := seeds.SeeddLines(ctx, db); err != nil {
		log.Fatalf("❌ Lines seeding failed:\n %v", err)
	}

	log.Println("✅ Seeding complete")
}
