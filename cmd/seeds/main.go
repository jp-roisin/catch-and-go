package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jp-roisin/catch-and-go/internal/database"
	"github.com/jp-roisin/catch-and-go/internal/database/seeds"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	service := database.New()
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
	log.Println("✅ Stops seeding successful")

	if err := seeds.SeedLines(ctx, db); err != nil {
		log.Fatalf("❌ Lines seeding failed:\n %v", err)
	}
	log.Println("✅ Lines seeding successful")

	if err := seeds.SeedLinesMetadatas(ctx, db); err != nil {
		log.Fatalf("❌ Lines metadata seeding failed:\n %v", err)
	}
	log.Println("✅ Lines metadata seeding successful")

	if err := seeds.SeedStopsByLines(ctx, db, service); err != nil {
		log.Fatalf("❌ Stops by Lines seeding failed:\n %v", err)
	}
	log.Println("✅ Stops by Lines seeding successful")

	if err := seeds.SeedLinesTextColors(ctx, db, service); err != nil {
		log.Fatalf("❌ Lines text colors seeding failed:\n %v", err)
	}
	log.Println("✅ Lines text colors seeding successful")

	log.Println("✅ Seeding complete")
}
