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

const batchSize = 100

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
