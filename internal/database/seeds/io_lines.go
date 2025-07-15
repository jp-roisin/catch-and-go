package seeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

const stopsByLineFilePath = "internal/database/seeds/data/stops-by-line-production.csv"

func IoSeedLines(ctx context.Context, db *sql.DB) error {
	linesResult, err := readCsvFile(stopsByLineFilePath)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO lines (code, destination, direction)
			VALUES (?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	count := 0

	for i, row := range linesResult {
		if i == 0 {
			continue // skip csv header
		}
		if len(row) != 4 {
			return fmt.Errorf("row %d has incorrect number of columns: %v", i+1, row)
		}

		destination := row[0]
		direction := row[1]
		code := row[2]

		var d i18nCell
		if err := json.Unmarshal([]byte(destination), &d); err != nil {
			return fmt.Errorf("invalid JSON in 'destination' at row %d: %v", i+1, err)
		}

		if !validString.MatchString(direction) {
			return fmt.Errorf("invalid stop ID at row %d: %q (must be alphanumeric)", i+1, direction)
		}

		if !validString.MatchString(code) {
			return fmt.Errorf("invalid stop ID at row %d: %q (must be alphanumeric)", i+1, code)
		}

		// Casting the direction ("City" || "Suburb") into a boolean
		booleanDirection := 1
		if direction == "Suburb" {
			booleanDirection = 0
		}
		_, err = stmt.ExecContext(ctx, code, destination, booleanDirection)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert row %d: %v", i+1, err)
		}

		count++

		if count%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return err
			}
			fmt.Printf("'Lines' batch complete: #%d \n", count/batchSize)

			tx, err = db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}

			stmt, err = tx.PrepareContext(ctx, `INSERT INTO lines (code, destination, direction) VALUES (?, ?, ?)`)
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
