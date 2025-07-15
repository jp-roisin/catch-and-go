package seeds

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Name struct {
	Fr string `json:"fr"`
	Nl string `json:"nl"`
}

var validStopID = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return records, nil
}

func IoSeedStops(ctx context.Context, db *sql.DB) error {
	stopsResult, err := readCsvFile("internal/database/seeds/data/stop-details-production.csv")
	if err != nil {
		return err
	}

	const batchSize = 100
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO stops (code, geo, name)
			VALUES (?, ?, ?)
		`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	count := 0

	for i, row := range stopsResult {
		if i == 0 {
			continue // skip the header
		}

		if len(row) != 3 {
			return fmt.Errorf("row %d has incorrect number of columns: %v", i+1, row)
		}
		location := row[0]
		code := row[1]
		name := row[2]

		var parsedLocation Location
		if err := json.Unmarshal([]byte(location), &parsedLocation); err != nil {
			return fmt.Errorf("invalid JSON in 'location' at row %d: %v", i+1, err)
		}

		if !validStopID.MatchString(code) {
			return fmt.Errorf("invalid stop ID at row %d: %q (must be alphanumeric)", i+1, code)
		}

		var parsedName Name
		if err := json.Unmarshal([]byte(name), &parsedName); err != nil {
			return fmt.Errorf("invalid JSON in 'name' at row %d: %v", i+1, err)
		}

		_, err = stmt.ExecContext(ctx, code, location, name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert row %d: %v", i+1, err)
		}

		count++

		if count%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return err
			}

			tx, err = db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}

			stmt, err = tx.PrepareContext(ctx, `INSERT INTO stops (code, geo, name) VALUES (?, ?, ?)`)
			if err != nil {
				tx.Rollback()
				return err
			}

			defer stmt.Close()
		}
	}

	// Commit remaining rows if any
	if count%batchSize != 0 {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
