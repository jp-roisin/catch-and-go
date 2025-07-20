package store

import (
	"encoding/json"
	"fmt"
)

const fallbackMode = "bus"
const fallbackColor = "#ccc"

type Direction int

const (
	TowardsSuburbs Direction = iota
	TowardsCity
)

type LineWithFallback struct {
	ID          int
	Code        string
	Destination string
	Direction   int
	Mode        string
	Color       string
}

func (l *Line) AddFallback() LineWithFallback {
	var mode, color string

	if l.Mode.Valid {
		mode = l.Mode.String
	} else {
		mode = fallbackMode
	}

	if l.Color.Valid {
		color = l.Color.String
	} else {
		color = fallbackColor
	}

	return LineWithFallback{
		ID:          int(l.ID),
		Code:        l.Code,
		Destination: l.Destination,
		Direction:   int(l.Direction),
		Mode:        mode,
		Color:       color,
	}
}

type i18n struct {
	Fr string `json:"fr"`
	Nl string `json:"nk"`
}

func (s *Stop) Translate(locale string) (Stop, error) {
	if locale != "fr" && locale != "nl" {
		return Stop{}, fmt.Errorf("%s is not a valid locale", locale)
	}
	var obj i18n
	if err := json.Unmarshal([]byte(s.Name), &obj); err != nil {
		return Stop{}, err
	}

	var translatedName string
	switch locale {
	case "fr":
		translatedName = obj.Fr
	case "nl":
		translatedName = obj.Nl
	}

	return Stop{
		ID:        s.ID,
		Code:      s.Code,
		Geo:       s.Geo,
		Name:      translatedName,
		CreatedAt: s.CreatedAt,
	}, nil
}
