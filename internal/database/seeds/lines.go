package seeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

const stopsByLineFilePath = "internal/database/seeds/data/stops-by-line-production.csv"

func SeedLines(ctx context.Context, db *sql.DB) error {
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

		if !validInteger.MatchString(code) {
			// Filtering lineIds like "N12"
			fmt.Printf("Skipping lineId containing letters: %s\n", code)
			continue
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

const shapeFilesPath = "internal/database/seeds/data/shapefiles-production.csv"

// SeedLinesMetadatas populates the `lines` table with additional metadata fields:
// - `mode`: represents the type of transportation ("bus", "tram", or "metro")
// - `color`: a HEX color code associated with the line (e.g., "#306196")
func SeedLinesMetadatas(ctx context.Context, db *sql.DB) error {
	metadataResult, err := readCsvFile(shapeFilesPath)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `UPDATE lines SET mode = ?, color = ? WHERE code = ?`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i, row := range metadataResult {
		if i == 0 {
			continue // skip csv header
		}
		if len(row) != 2 {
			return fmt.Errorf("row %d has incorrect number of columns: %v", i+1, row)
		}

		// Parse the first CSV column to extract the line ID and transportation mode.
		// Example: "002m" yields:
		// - lineId: 2   (parsed as int, with leading zeros removed)
		// - mode:   "metro" (derived from the final character)
		idWithMode := validLineIdWithMode.FindStringSubmatch(row[0])
		lineId, err := strconv.Atoi(idWithMode[1])
		if err != nil {
			return fmt.Errorf("invalid stop ID at row %d: %d (must be valid integer)", i+1, lineId)
		}
		mode, ok := modeMap[idWithMode[2]]
		if !ok {
			return fmt.Errorf("invalid mode at row %d: %s (must be either `m`, `b` or `t`)", i+1, mode)
		}

		color := row[1]
		if !validHexColor.MatchString(color) {
			return fmt.Errorf("invalid hex color at row %d: %s", i+1, color)
		}

		_, err = stmt.ExecContext(ctx, mode, color, lineId)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert row %d: %v", i+1, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
