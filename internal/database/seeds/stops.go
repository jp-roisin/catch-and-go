package seeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type apiResponse struct {
	TotalCount int    `json:"total_count"`
	Results    []stop `json:"results"`
}

type stop struct {
	GPSCoordinates string `json:"gpscoordinates"`
	ID             string `json:"id"`
	Name           string `json:"name"`
}

const (
	baseURL = "https://data.stib-mivb.brussels/api/explore/v2.1/catalog/datasets/stop-details-production/records"
	batch   = 100
)

// SeedStops fetches all stops from the external API and stores them in the DB in batches.
func SeedStops(ctx context.Context, db *sql.DB) error {
	offset := 0
	totalCount := -1

	db.ExecContext(ctx, "DELETE FROM stops;")
	db.ExecContext(ctx, "DELETE FROM sqlite_sequence WHERE name = 'stops';")
	log.Println("✅ Done clearing stops")

	for {
		url := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, batch, offset)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		apiKey := os.Getenv("STIB_API_KEY")
		req.Header.Set("Authorization", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			return fmt.Errorf("failed to fetch stops: %w", err)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		var data apiResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		if totalCount == -1 {
			totalCount = data.TotalCount
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO stops (code, geo, name)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("prepare failed: %w", err)
		}

		for _, s := range data.Results {
			if _, err := stmt.ExecContext(ctx, s.ID, s.GPSCoordinates, s.Name); err != nil {
				stmt.Close()
				tx.Rollback()
				return fmt.Errorf("insert failed: %w", err)
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}

		log.Printf("Seeded stops batch offset=%d", offset)
		offset += batch

		if offset >= totalCount {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Println("✅ Done seeding stops")
	return nil
}
