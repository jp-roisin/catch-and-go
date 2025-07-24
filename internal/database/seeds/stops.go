package seeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

const stopDetailsFilePath = "internal/database/seeds/data/stop-details-production.csv"

func SeedStops(ctx context.Context, db *sql.DB) error {
	stopsResult, err := readCsvFile(stopDetailsFilePath)
	if err != nil {
		return err
	}

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

	// Insert a fallback "unknown stop" as the first row in the stops table.
	// This stop is used as a default reference in the stops_by_lines table
	// when a stop cannot be resolved from the csv.
	us, err := getUnknownStop()
	if err != nil {
		return fmt.Errorf("failed to generate the unknown stop: %v", err)
	}

	_, err = stmt.ExecContext(ctx, us.Code, us.Geo, us.Name)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert unknown stop row: %v", err)
	}

	count := 0

	for i, row := range stopsResult {
		if i == 0 {
			continue // skip csv header
		}

		if len(row) != 3 {
			return fmt.Errorf("row %d has incorrect number of columns: %v", i+1, row)
		}
		location := row[0]
		code := row[1]
		name := row[2]

		var l Location
		if err := json.Unmarshal([]byte(location), &l); err != nil {
			return fmt.Errorf("invalid JSON in 'location' at row %d: %v", i+1, err)
		}

		// TODO: migrate the db to make the column int64
		codeInt, err := strconv.Atoi(removeTrailingLetter(code))
		if err != nil {
			return fmt.Errorf("invalid code at row %d: %s (must be valid integer)", i+1, code)
		}

		if !validString.MatchString(code) {
			return fmt.Errorf("invalid stop ID at row %d: %q (must be alphanumeric)", i+1, code)
		}

		var n i18nCell
		if err := json.Unmarshal([]byte(name), &n); err != nil {
			return fmt.Errorf("invalid JSON in 'name' at row %d: %v", i+1, err)
		}

		_, err = stmt.ExecContext(ctx, codeInt, location, name)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert row %d: %v", i+1, err)
		}

		count++

		if count%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return err
			}
			fmt.Printf("'Stops' batch complete: #%d \n", count/batchSize)

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
