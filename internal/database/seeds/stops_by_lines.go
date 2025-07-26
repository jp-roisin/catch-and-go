package seeds

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jp-roisin/catch-and-go/internal/database"
	"github.com/jp-roisin/catch-and-go/internal/database/store"
)

type lineStop struct {
	Order int    `json:"order"`
	Code  string `json:"id"`
}

func SeedStopsByLines(ctx context.Context, db *sql.DB, service database.Service) error {
	linesResult, err := readCsvFile(stopsByLineFilePath)
	if err != nil {
		return err
	}

	for i, row := range linesResult {
		if i == 0 {
			continue // skip csv header
		}
		if len(row) != 4 {
			return fmt.Errorf("row %d has incorrect number of columns: %v", i+1, row)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, `INSERT INTO stops_by_lines (stop_id, line_id, "order") VALUES (?, ?, ?)`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		direction := row[1]
		code := row[2]      // lineCode
		lineStops := row[3] // stopCode

		// Filtering out lineIds (like "N12" which is the Noctis / night shift for the line)
		// that have not been inserted in `lines`
		codeInt, err := strconv.Atoi(removeTrailingLetter(code))
		if err != nil {
			continue
		}

		var ls []lineStop
		if err := json.Unmarshal([]byte(lineStops), &ls); err != nil {
			return fmt.Errorf("invalid JSON in 'line_stops' at row %d: %v", i+1, err)
		}

		line, err := service.GetLine(ctx, store.GetLineParams{
			Code:      strconv.FormatInt(int64(codeInt), 10),
			Direction: int64(directionToBoolean(direction)),
		})
		if err != nil {
			return fmt.Errorf("Line ID: %s was not found in lines at row %d: %v", code, i+1, err)
		}

		for _, st := range ls {
			var stop store.Stop
			sanitazedStopId, err := strconv.Atoi(removeTrailingLetter(st.Code))
			if err != nil {
				stop.ID = 1 // Fallback to the first row inserted in `stops` (see getUnknownStop())
			}
			stop, err = service.GetStop(ctx, strconv.FormatInt(int64(sanitazedStopId), 10))
			if err != nil {
				stop.ID = 1 // Fallback to the first row inserted in `stops` (see getUnknownStop())
			}

			_, err = stmt.ExecContext(ctx, stop.ID, line.ID, st.Order)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert row %d: %v", i+1, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit: %v", err)
		}

	}

	return nil
}
