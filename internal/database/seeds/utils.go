package seeds

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
)

type i18nCell struct {
	Fr string `json:"fr"`
	Nl string `json:"nl"`
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
