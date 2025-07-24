package externalapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type I18n struct {
	FR string `json:"fr"`
	NL string `json:"nl"`
}

type PassingTime struct {
	Destination         I18n   `json:"destination"`
	ExpectedArrivalTime string `json:"expectedArrivalTime"`
	LineID              string `json:"lineId"`
}

// Custom type to handle JSON-encoded string
type PassingTimeList []PassingTime

func (p *PassingTimeList) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), (*[]PassingTime)(p))
}

type WaitingTime struct {
	PointID      string          `json:"pointid"`
	LineID       string          `json:"lineid"`
	PassingTimes PassingTimeList `json:"passingtimes"`
}

type Response struct {
	TotalCount   int           `json:"total_count"`
	WaitingTimes []WaitingTime `json:"results"`
}

const baseUrl = "https://data.stib-mivb.brussels/api/explore/v2.1/catalog/datasets/waiting-time-rt-production/records"

func GetWaitingTimeForStop(stopCode string) (Response, error) {
	var result Response
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?where=pointid=%s", baseUrl, stopCode), nil)
	if err != nil {
		return result, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Apikey %s", os.Getenv("STIB_API_KEY")))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return result, fmt.Errorf("server error: %s", string(body))
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("json decode failed: %w", err)
	}

	return result, nil
}
