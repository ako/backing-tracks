package parser

import (
	"math"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Track represents a complete backing track
type Track struct {
	Info        TrackInfo        `yaml:"track"`
	Progression ChordProgression `yaml:"chord_progression"`
	Rhythm      *Rhythm          `yaml:"rhythm,omitempty"`
	Bass        *Bass            `yaml:"bass,omitempty"`
	Drums       *Drums           `yaml:"drums,omitempty"`
}

// TrackInfo contains metadata about the track
type TrackInfo struct {
	Title         string `yaml:"title"`
	Key           string `yaml:"key"`
	Tempo         int    `yaml:"tempo"`
	TimeSignature string `yaml:"time_signature"`
	Style         string `yaml:"style"`
}

// ChordProgression represents the chord sequence
type ChordProgression struct {
	Pattern       string `yaml:"pattern"`
	BarsPerChord  int    `yaml:"bars_per_chord"`
	Repeat        int    `yaml:"repeat"`
}

// Chord represents a single chord with duration
type Chord struct {
	Symbol string
	Bars   float64  // Supports fractional bars (0.5, 1.5, 2.0, etc.)
}

// LoadTrack reads and parses a BTML file
func LoadTrack(filename string) (*Track, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var track Track
	if err := yaml.Unmarshal(data, &track); err != nil {
		return nil, err
	}

	// Set defaults
	if track.Progression.BarsPerChord == 0 {
		track.Progression.BarsPerChord = 1
	}
	if track.Progression.Repeat == 0 {
		track.Progression.Repeat = 1
	}

	return &track, nil
}

// GetChords parses the pattern string and returns a slice of chords
// Supports inline duration notation: "Em*2" = Em for 2 bars, "G*0.5" = G for half a bar
func (cp *ChordProgression) GetChords() []Chord {
	parts := strings.Fields(cp.Pattern)
	chords := make([]Chord, 0, len(parts))

	for _, part := range parts {
		symbol, bars := parseChordWithDuration(part, cp.BarsPerChord)
		chords = append(chords, Chord{
			Symbol: symbol,
			Bars:   bars,
		})
	}

	// Apply repeat
	if cp.Repeat > 1 {
		original := chords
		for i := 1; i < cp.Repeat; i++ {
			chords = append(chords, original...)
		}
	}

	return chords
}

// parseChordWithDuration extracts chord symbol and duration
// Supports: "Em*2" (2 bars), "G*1" (1 bar), "C*0.5" (half bar), "D" (default bars)
func parseChordWithDuration(part string, defaultBars int) (string, float64) {
	// Check for duration notation: ChordSymbol*Duration
	if idx := strings.Index(part, "*"); idx != -1 {
		symbol := part[:idx]
		durationStr := part[idx+1:]

		// Parse duration as float (to support 0.5, 1.5, etc.)
		if duration, err := strconv.ParseFloat(durationStr, 64); err == nil {
			// Return the exact duration (supports fractions!)
			if duration <= 0 {
				duration = 0.5 // Minimum half bar
			}
			return symbol, duration
		}
	}

	// No duration specified, use default
	return part, float64(defaultBars)
}

// TotalBars calculates the total number of bars in the progression
func (cp *ChordProgression) TotalBars() int {
	chords := cp.GetChords()
	total := 0.0
	for _, chord := range chords {
		total += chord.Bars
	}
	return int(math.Ceil(total))
}

// Bass represents the bass line configuration
type Bass struct {
	Style   string  `yaml:"style"`            // walking, root, root_fifth, etc.
	Pattern string  `yaml:"pattern,omitempty"` // Custom pattern (optional)
	Swing   float64 `yaml:"swing,omitempty"`   // Swing feel (0.5 = straight, 0.67 = triplet)
}

// Rhythm represents the chord strumming/voicing pattern
type Rhythm struct {
	Style   string  `yaml:"style"`             // whole, half, quarter, eighth, strum_down, strum_up_down, folk, shuffle_strum, pattern
	Pattern string  `yaml:"pattern,omitempty"` // Custom pattern: D=down, U=up, .=rest, x=muted, e.g. "D.DU.UDU"
	Swing   float64 `yaml:"swing,omitempty"`   // Swing feel (0.5 = straight, 0.67 = triplet)
	Accent  string  `yaml:"accent,omitempty"`  // Which beats to accent: "1", "1,3", "2,4", etc.
}

// Drums represents the drum configuration
type Drums struct {
	Style    string          `yaml:"style"`    // shuffle, rock_beat, jazz_swing, etc.
	Kick     *DrumPattern    `yaml:"kick,omitempty"`
	Snare    *DrumPattern    `yaml:"snare,omitempty"`
	Hihat    *DrumPattern    `yaml:"hihat,omitempty"`
	Ride     *DrumPattern    `yaml:"ride,omitempty"`
	Intensity float64        `yaml:"intensity,omitempty"` // 0.0 to 1.0
}

// DrumPattern represents a drum pattern (can be Euclidean or explicit)
type DrumPattern struct {
	// Option 1: Euclidean rhythm
	Euclidean *EuclideanRhythm `yaml:"euclidean,omitempty"`

	// Option 2: Explicit pattern string
	Pattern string `yaml:"pattern,omitempty"`

	// Option 3: Explicit beat positions
	Beats []int `yaml:"beats,omitempty"`
}

// EuclideanRhythm defines an algorithmic rhythm pattern
type EuclideanRhythm struct {
	Hits     int `yaml:"hits"`      // Number of hits
	Steps    int `yaml:"steps"`     // Total steps
	Rotation int `yaml:"rotation"`  // Rotation offset
}
