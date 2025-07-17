package seeds

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/jp-roisin/catch-and-go/internal/database/store"
)

type i18nCell struct {
	Fr string `json:"fr"`
	Nl string `json:"nl"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var validString = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var validInteger = regexp.MustCompile(`^\d+$`)
var validHexColor = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)
var validLineIdWithMode = regexp.MustCompile(`^(\d+)([a-zA-Z])$`)

const batchSize = 100

var modeMap = map[string]string{
	"m": "metro",
	"b": "bus",
	"t": "tram",
}

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

// Converts a direction string ("City" or "Suburb") to an integer
// representation suitable for SQLite, which doesn't support boolean types.
// Returns 1 for "City" and 0 for "Suburb".
func directionToBoolean(direction string) int {
	if direction == "Suburb" {
		return 0
	}
	return 1
}

// Returns a fallback Stop struct representing an "unknown stop" location.
// The coordinates are set to the Brussels Grand Place as a neutral central location.
// The stop name is localized in French and Dutch with a clear "not found" label.
func getUnknownStop() (store.Stop, error) {
	bxl, err := json.Marshal(Location{
		Latitude:  50.8468, // Brussels Grand Place latitude
		Longitude: 4.3524,  // Brussels Grand Place longitude
	})
	if err != nil {
		return store.Stop{}, err
	}

	unknownName, err := json.Marshal(i18nCell{
		Fr: "ARRÊT NON TROUVÉ",
		Nl: "STOP NIET GEVONDEN",
	})
	if err != nil {
		return store.Stop{}, err
	}

	return store.Stop{
		Code: "0001",
		Geo:  string(bxl),
		Name: string(unknownName),
	}, nil
}
