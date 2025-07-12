package seeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type linesApiResponse struct {
	TotalCount int    `json:"total_count"`
	Results    []line `json:"results"`
}

type line struct {
	Destination string `json:"destination"`
	Direction   string `json:"direction"`
	LineId      string `json:"lineid"`
}

var lineQuery = struct {
	baseURL string
	batch   int
}{
	baseURL: "https://data.stib-mivb.brussels/api/explore/v2.1/catalog/datasets/stops-by-line-production/records",
	batch:   100,
}

func SeeddLines(ctx context.Context, db *sql.DB) error {
	offset := 0
	totalCount := -1

	// db.ExecContext(ctx, "DELETE FROM lines;")
	// db.ExecContext(ctx, "DELETE FROM sqlite_sequence WHERE name = 'lines';")
	// log.Println("✅ Done clearing lines")

	for {
		url := fmt.Sprintf("%s?limit=%d&offset=%d", lineQuery.baseURL, lineQuery.batch, offset)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			return fmt.Errorf("failed to fetch lines: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("💥HTTP 'stops-by-line' request failed with status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		var data linesApiResponse
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
			INSERT INTO lines (code, destination, direction)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("prepare failed: %w", err)
		}

		for _, l := range data.Results {
			// Casting the direction ("City" || "Suburb") into a boolean
			booleanDirection := 1
			if l.Direction == "Suburb" {
				booleanDirection = 0
			}

			if _, err := stmt.ExecContext(ctx, l.LineId, l.Destination, booleanDirection); err != nil {
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
		offset += stopQuery.batch

		if offset >= totalCount {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Println("✅ Done seeding lines")
	return nil
}
