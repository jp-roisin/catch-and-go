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
	TextColor   string
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
		TextColor:   l.TextColor,
	}
}

type i18n struct {
	Fr string `json:"fr"`
	Nl string `json:"nl"`
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

func (d *ListDashboardsFromSessionRow) Translate(locale string) (ListDashboardsFromSessionRow, error) {
	if locale != "fr" && locale != "nl" {
		return ListDashboardsFromSessionRow{}, fmt.Errorf("%s is not a valid locale", locale)
	}
	var obj i18n
	if err := json.Unmarshal([]byte(d.StopName), &obj); err != nil {
		return ListDashboardsFromSessionRow{}, err
	}

	var translatedName string
	switch locale {
	case "fr":
		translatedName = obj.Fr
	case "nl":
		translatedName = obj.Nl
	}

	return ListDashboardsFromSessionRow{
		DashboardID:        d.DashboardID,
		SessionID:          d.SessionID,
		StopID:             d.StopID,
		DashboardCreatedAt: d.DashboardCreatedAt,
		StopID_2:           d.StopID_2,
		StopCode:           d.StopCode,
		StopGeo:            d.StopGeo,
		StopName:           translatedName,
		StopCreatedAt:      d.StopCreatedAt,
	}, nil
}

func (l *Line) Translate(locale string) (Line, error) {
	var line Line
	if locale != "fr" && locale != "nl" {
		return line, fmt.Errorf("%s is not a valid locale", locale)
	}
	var obj i18n
	if err := json.Unmarshal([]byte(l.Destination), &obj); err != nil {
		return line, err
	}

	var translatedDestination string
	switch locale {
	case "fr":
		translatedDestination = obj.Fr
	case "nl":
		translatedDestination = obj.Nl
	}

	return Line{
		ID:          l.ID,
		Code:        l.Code,
		Direction:   l.Direction,
		Destination: translatedDestination,
		CreatedAt:   l.CreatedAt,
	}, nil
}
