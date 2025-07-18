package store

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

func (s *Line) AddFallback() LineWithFallback {
	var mode, color string

	if s.Mode.Valid {
		mode = s.Mode.String
	} else {
		mode = fallbackMode
	}

	if s.Color.Valid {
		color = s.Color.String
	} else {
		color = fallbackColor
	}

	return LineWithFallback{
		ID:          int(s.ID),
		Code:        s.Code,
		Destination: s.Destination,
		Direction:   int(s.Direction),
		Mode:        mode,
		Color:       color,
	}
}
