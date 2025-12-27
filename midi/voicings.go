package midi

import (
	"backing-tracks/theory"
)

// GuitarVoicing represents a chord shape on guitar for tablature display
// Frets: -1 = muted/not played, 0 = open, 1+ = fret number
// Strings are indexed 0-5 where 0 = low E, 5 = high e
type GuitarVoicing struct {
	Name       string
	Frets      [6]int  // Low E, A, D, G, B, high e
	Fingers    [6]int  // 0 = not used, 1-4 = finger number, 5 = thumb
	BassFret   int     // Which fret has the bass note (for reference)
	BassString int     // Which string is the bass (0-5)
}

// GetFretNote returns the MIDI note for a given string and fret using a tuning
func GetFretNote(tuning theory.Tuning, stringNum int, fret int) int {
	if fret < 0 || stringNum < 0 || stringNum >= len(tuning.Notes) {
		return -1 // muted
	}
	return tuning.Notes[stringNum] + fret
}

// GetFretNoteWithCapo returns the MIDI note adjusted for capo position
func GetFretNoteWithCapo(tuning theory.Tuning, stringNum int, fret int, capo int) int {
	if fret < 0 || stringNum < 0 || stringNum >= len(tuning.Notes) {
		return -1 // muted
	}
	// Capo raises the pitch, so add capo to the note
	return tuning.Notes[stringNum] + fret + capo
}

// GuitarVoicings contains common guitar chord voicings (in standard tuning, no capo)
// These are the "shapes" that can be transposed with capo
var GuitarVoicings = map[string]GuitarVoicing{
	// Major chords
	"C": {
		Name:       "C",
		Frets:      [6]int{-1, 3, 2, 0, 1, 0},
		Fingers:    [6]int{0, 3, 2, 0, 1, 0},
		BassFret:   3,
		BassString: 1,
	},
	"D": {
		Name:       "D",
		Frets:      [6]int{-1, -1, 0, 2, 3, 2},
		Fingers:    [6]int{0, 0, 0, 1, 3, 2},
		BassFret:   0,
		BassString: 2,
	},
	"E": {
		Name:       "E",
		Frets:      [6]int{0, 2, 2, 1, 0, 0},
		Fingers:    [6]int{0, 2, 3, 1, 0, 0},
		BassFret:   0,
		BassString: 0,
	},
	"F": {
		Name:       "F",
		Frets:      [6]int{1, 3, 3, 2, 1, 1},
		Fingers:    [6]int{1, 3, 4, 2, 1, 1},
		BassFret:   1,
		BassString: 0,
	},
	"G": {
		Name:       "G",
		Frets:      [6]int{3, 2, 0, 0, 0, 3},
		Fingers:    [6]int{2, 1, 0, 0, 0, 3},
		BassFret:   3,
		BassString: 0,
	},
	"A": {
		Name:       "A",
		Frets:      [6]int{-1, 0, 2, 2, 2, 0},
		Fingers:    [6]int{0, 0, 1, 2, 3, 0},
		BassFret:   0,
		BassString: 1,
	},
	"B": {
		Name:       "B",
		Frets:      [6]int{-1, 2, 4, 4, 4, 2},
		Fingers:    [6]int{0, 1, 2, 3, 4, 1},
		BassFret:   2,
		BassString: 1,
	},

	// Minor chords
	"Am": {
		Name:       "Am",
		Frets:      [6]int{-1, 0, 2, 2, 1, 0},
		Fingers:    [6]int{0, 0, 2, 3, 1, 0},
		BassFret:   0,
		BassString: 1,
	},
	"Bm": {
		Name:       "Bm",
		Frets:      [6]int{-1, 2, 4, 4, 3, 2},
		Fingers:    [6]int{0, 1, 3, 4, 2, 1},
		BassFret:   2,
		BassString: 1,
	},
	"Cm": {
		Name:       "Cm",
		Frets:      [6]int{-1, 3, 5, 5, 4, 3},
		Fingers:    [6]int{0, 1, 3, 4, 2, 1},
		BassFret:   3,
		BassString: 1,
	},
	"Dm": {
		Name:       "Dm",
		Frets:      [6]int{-1, -1, 0, 2, 3, 1},
		Fingers:    [6]int{0, 0, 0, 2, 3, 1},
		BassFret:   0,
		BassString: 2,
	},
	"Em": {
		Name:       "Em",
		Frets:      [6]int{0, 2, 2, 0, 0, 0},
		Fingers:    [6]int{0, 2, 3, 0, 0, 0},
		BassFret:   0,
		BassString: 0,
	},
	"Fm": {
		Name:       "Fm",
		Frets:      [6]int{1, 3, 3, 1, 1, 1},
		Fingers:    [6]int{1, 3, 4, 1, 1, 1},
		BassFret:   1,
		BassString: 0,
	},
	"Gm": {
		Name:       "Gm",
		Frets:      [6]int{3, 5, 5, 3, 3, 3},
		Fingers:    [6]int{1, 3, 4, 1, 1, 1},
		BassFret:   3,
		BassString: 0,
	},

	// Seventh chords
	"A7": {
		Name:       "A7",
		Frets:      [6]int{-1, 0, 2, 0, 2, 0},
		Fingers:    [6]int{0, 0, 1, 0, 2, 0},
		BassFret:   0,
		BassString: 1,
	},
	"B7": {
		Name:       "B7",
		Frets:      [6]int{-1, 2, 1, 2, 0, 2},
		Fingers:    [6]int{0, 2, 1, 3, 0, 4},
		BassFret:   2,
		BassString: 1,
	},
	"C7": {
		Name:       "C7",
		Frets:      [6]int{-1, 3, 2, 3, 1, 0},
		Fingers:    [6]int{0, 3, 2, 4, 1, 0},
		BassFret:   3,
		BassString: 1,
	},
	"D7": {
		Name:       "D7",
		Frets:      [6]int{-1, -1, 0, 2, 1, 2},
		Fingers:    [6]int{0, 0, 0, 2, 1, 3},
		BassFret:   0,
		BassString: 2,
	},
	"E7": {
		Name:       "E7",
		Frets:      [6]int{0, 2, 0, 1, 0, 0},
		Fingers:    [6]int{0, 2, 0, 1, 0, 0},
		BassFret:   0,
		BassString: 0,
	},
	"F7": {
		Name:       "F7",
		Frets:      [6]int{1, 3, 1, 2, 1, 1},
		Fingers:    [6]int{1, 3, 1, 2, 1, 1},
		BassFret:   1,
		BassString: 0,
	},
	"G7": {
		Name:       "G7",
		Frets:      [6]int{3, 2, 0, 0, 0, 1},
		Fingers:    [6]int{3, 2, 0, 0, 0, 1},
		BassFret:   3,
		BassString: 0,
	},

	// Minor seventh chords
	"Am7": {
		Name:       "Am7",
		Frets:      [6]int{-1, 0, 2, 0, 1, 0},
		Fingers:    [6]int{0, 0, 2, 0, 1, 0},
		BassFret:   0,
		BassString: 1,
	},
	"Bm7": {
		Name:       "Bm7",
		Frets:      [6]int{-1, 2, 4, 2, 3, 2},
		Fingers:    [6]int{0, 1, 3, 1, 2, 1},
		BassFret:   2,
		BassString: 1,
	},
	"Cm7": {
		Name:       "Cm7",
		Frets:      [6]int{-1, 3, 5, 3, 4, 3},
		Fingers:    [6]int{0, 1, 3, 1, 2, 1},
		BassFret:   3,
		BassString: 1,
	},
	"Dm7": {
		Name:       "Dm7",
		Frets:      [6]int{-1, -1, 0, 2, 1, 1},
		Fingers:    [6]int{0, 0, 0, 2, 1, 1},
		BassFret:   0,
		BassString: 2,
	},
	"Em7": {
		Name:       "Em7",
		Frets:      [6]int{0, 2, 0, 0, 0, 0},
		Fingers:    [6]int{0, 2, 0, 0, 0, 0},
		BassFret:   0,
		BassString: 0,
	},
	"Fm7": {
		Name:       "Fm7",
		Frets:      [6]int{1, 3, 1, 1, 1, 1},
		Fingers:    [6]int{1, 3, 1, 1, 1, 1},
		BassFret:   1,
		BassString: 0,
	},
	"Gm7": {
		Name:       "Gm7",
		Frets:      [6]int{3, 5, 3, 3, 3, 3},
		Fingers:    [6]int{1, 3, 1, 1, 1, 1},
		BassFret:   3,
		BassString: 0,
	},

	// Major seventh chords
	"Amaj7": {
		Name:       "Amaj7",
		Frets:      [6]int{-1, 0, 2, 1, 2, 0},
		Fingers:    [6]int{0, 0, 2, 1, 3, 0},
		BassFret:   0,
		BassString: 1,
	},
	"Bmaj7": {
		Name:       "Bmaj7",
		Frets:      [6]int{-1, 2, 4, 3, 4, 2},
		Fingers:    [6]int{0, 1, 3, 2, 4, 1},
		BassFret:   2,
		BassString: 1,
	},
	"Cmaj7": {
		Name:       "Cmaj7",
		Frets:      [6]int{-1, 3, 2, 0, 0, 0},
		Fingers:    [6]int{0, 3, 2, 0, 0, 0},
		BassFret:   3,
		BassString: 1,
	},
	"Dmaj7": {
		Name:       "Dmaj7",
		Frets:      [6]int{-1, -1, 0, 2, 2, 2},
		Fingers:    [6]int{0, 0, 0, 1, 1, 1},
		BassFret:   0,
		BassString: 2,
	},
	"Emaj7": {
		Name:       "Emaj7",
		Frets:      [6]int{0, 2, 1, 1, 0, 0},
		Fingers:    [6]int{0, 3, 1, 2, 0, 0},
		BassFret:   0,
		BassString: 0,
	},
	"Fmaj7": {
		Name:       "Fmaj7",
		Frets:      [6]int{-1, -1, 3, 2, 1, 0},
		Fingers:    [6]int{0, 0, 3, 2, 1, 0},
		BassFret:   3,
		BassString: 2,
	},
	"Gmaj7": {
		Name:       "Gmaj7",
		Frets:      [6]int{3, 2, 0, 0, 0, 2},
		Fingers:    [6]int{2, 1, 0, 0, 0, 3},
		BassFret:   3,
		BassString: 0,
	},

	// Suspended chords
	"Asus2": {
		Name:       "Asus2",
		Frets:      [6]int{-1, 0, 2, 2, 0, 0},
		Fingers:    [6]int{0, 0, 1, 2, 0, 0},
		BassFret:   0,
		BassString: 1,
	},
	"Asus4": {
		Name:       "Asus4",
		Frets:      [6]int{-1, 0, 2, 2, 3, 0},
		Fingers:    [6]int{0, 0, 1, 2, 3, 0},
		BassFret:   0,
		BassString: 1,
	},
	"Dsus2": {
		Name:       "Dsus2",
		Frets:      [6]int{-1, -1, 0, 2, 3, 0},
		Fingers:    [6]int{0, 0, 0, 1, 2, 0},
		BassFret:   0,
		BassString: 2,
	},
	"Dsus4": {
		Name:       "Dsus4",
		Frets:      [6]int{-1, -1, 0, 2, 3, 3},
		Fingers:    [6]int{0, 0, 0, 1, 2, 3},
		BassFret:   0,
		BassString: 2,
	},
	"Esus4": {
		Name:       "Esus4",
		Frets:      [6]int{0, 2, 2, 2, 0, 0},
		Fingers:    [6]int{0, 1, 2, 3, 0, 0},
		BassFret:   0,
		BassString: 0,
	},
	"Gsus4": {
		Name:       "Gsus4",
		Frets:      [6]int{3, 3, 0, 0, 1, 3},
		Fingers:    [6]int{2, 3, 0, 0, 1, 4},
		BassFret:   3,
		BassString: 0,
	},

	// Add9 chords
	"Cadd9": {
		Name:       "Cadd9",
		Frets:      [6]int{-1, 3, 2, 0, 3, 0},
		Fingers:    [6]int{0, 2, 1, 0, 3, 0},
		BassFret:   3,
		BassString: 1,
	},
	"Dadd9": {
		Name:       "Dadd9",
		Frets:      [6]int{-1, -1, 0, 2, 3, 0},
		Fingers:    [6]int{0, 0, 0, 1, 2, 0},
		BassFret:   0,
		BassString: 2,
	},
	"Eadd9": {
		Name:       "Eadd9",
		Frets:      [6]int{0, 2, 2, 1, 0, 2},
		Fingers:    [6]int{0, 2, 3, 1, 0, 4},
		BassFret:   0,
		BassString: 0,
	},
	"Gadd9": {
		Name:       "Gadd9",
		Frets:      [6]int{3, 2, 0, 2, 0, 3},
		Fingers:    [6]int{2, 1, 0, 3, 0, 4},
		BassFret:   3,
		BassString: 0,
	},

	// Sharp/flat variants
	"F#": {
		Name:       "F#",
		Frets:      [6]int{2, 4, 4, 3, 2, 2},
		Fingers:    [6]int{1, 3, 4, 2, 1, 1},
		BassFret:   2,
		BassString: 0,
	},
	"F#m": {
		Name:       "F#m",
		Frets:      [6]int{2, 4, 4, 2, 2, 2},
		Fingers:    [6]int{1, 3, 4, 1, 1, 1},
		BassFret:   2,
		BassString: 0,
	},
	"F#m7": {
		Name:       "F#m7",
		Frets:      [6]int{2, 4, 2, 2, 2, 2},
		Fingers:    [6]int{1, 3, 1, 1, 1, 1},
		BassFret:   2,
		BassString: 0,
	},
	"F#7": {
		Name:       "F#7",
		Frets:      [6]int{2, 4, 2, 3, 2, 2},
		Fingers:    [6]int{1, 3, 1, 2, 1, 1},
		BassFret:   2,
		BassString: 0,
	},
	"Bb": {
		Name:       "Bb",
		Frets:      [6]int{-1, 1, 3, 3, 3, 1},
		Fingers:    [6]int{0, 1, 2, 3, 4, 1},
		BassFret:   1,
		BassString: 1,
	},
	"Bbm": {
		Name:       "Bbm",
		Frets:      [6]int{-1, 1, 3, 3, 2, 1},
		Fingers:    [6]int{0, 1, 3, 4, 2, 1},
		BassFret:   1,
		BassString: 1,
	},
	"Eb": {
		Name:       "Eb",
		Frets:      [6]int{-1, -1, 1, 3, 4, 3},
		Fingers:    [6]int{0, 0, 1, 2, 4, 3},
		BassFret:   1,
		BassString: 2,
	},
	"Ab": {
		Name:       "Ab",
		Frets:      [6]int{4, 6, 6, 5, 4, 4},
		Fingers:    [6]int{1, 3, 4, 2, 1, 1},
		BassFret:   4,
		BassString: 0,
	},
	"C#m": {
		Name:       "C#m",
		Frets:      [6]int{-1, 4, 6, 6, 5, 4},
		Fingers:    [6]int{0, 1, 3, 4, 2, 1},
		BassFret:   4,
		BassString: 1,
	},
	"C#m7": {
		Name:       "C#m7",
		Frets:      [6]int{-1, 4, 6, 4, 5, 4},
		Fingers:    [6]int{0, 1, 3, 1, 2, 1},
		BassFret:   4,
		BassString: 1,
	},
	"G#m": {
		Name:       "G#m",
		Frets:      [6]int{4, 6, 6, 4, 4, 4},
		Fingers:    [6]int{1, 3, 4, 1, 1, 1},
		BassFret:   4,
		BassString: 0,
	},
	"F#sus4": {
		Name:       "F#sus4",
		Frets:      [6]int{2, 4, 4, 4, 2, 2},
		Fingers:    [6]int{1, 2, 3, 4, 1, 1},
		BassFret:   2,
		BassString: 0,
	},
}

// GetGuitarVoicing returns the guitar voicing for a chord symbol
// Uses the predefined voicing if available, otherwise generates one dynamically
func GetGuitarVoicing(symbol string, tuning theory.Tuning) GuitarVoicing {
	// First try exact match in predefined voicings
	if voicing, ok := GuitarVoicings[symbol]; ok {
		return voicing
	}

	// Normalize and try again
	normalized := normalizeChordSymbol(symbol)
	if voicing, ok := GuitarVoicings[normalized]; ok {
		return voicing
	}

	// Use theory package's dynamic generation and convert to GuitarVoicing
	theoryVoicing := theory.GenerateChordVoicing(symbol, tuning)
	return convertTheoryVoicing(symbol, theoryVoicing)
}

// GetGuitarVoicingWithCapo returns voicing adjusted for capo
// The frets in the voicing are relative to the capo position
func GetGuitarVoicingWithCapo(symbol string, tuning theory.Tuning, capo int) GuitarVoicing {
	voicing := GetGuitarVoicing(symbol, tuning)

	// Adjust frets relative to capo
	// If capo is on fret 2, and voicing says fret 0, that's actually fret 2
	// For display purposes, we show frets relative to capo
	// But for MIDI note calculation, we add capo offset

	return voicing // Shape stays the same, actual pitch is calculated with capo offset
}

// convertTheoryVoicing converts theory.ChordVoicing to GuitarVoicing
func convertTheoryVoicing(symbol string, tv theory.ChordVoicing) GuitarVoicing {
	gv := GuitarVoicing{
		Name:       symbol,
		Frets:      tv.Frets,
		Fingers:    [6]int{0, 0, 0, 0, 0, 0}, // Default - no finger info
		BassFret:   tv.BaseFret,
		BassString: 0,
	}

	// Find bass string (lowest non-muted string)
	for i := 0; i < 6; i++ {
		if tv.Frets[i] >= 0 {
			gv.BassString = i
			gv.BassFret = tv.Frets[i]
			break
		}
	}

	return gv
}

// normalizeChordSymbol handles common chord symbol variations
func normalizeChordSymbol(symbol string) string {
	// Handle common aliases
	replacements := map[string]string{
		"min":   "m",
		"minor": "m",
		"M7":    "maj7",
		"Δ":     "maj7",
		"Δ7":    "maj7",
		"-":     "m",
		"-7":    "m7",
		"+":     "aug",
		"°":     "dim",
		"o":     "dim",
		"o7":    "dim7",
	}

	result := symbol
	root := parseRootFromSymbol(result)
	if len(root) < len(result) {
		suffix := result[len(root):]
		if replacement, ok := replacements[suffix]; ok {
			result = root + replacement
		}
	}

	return result
}

// parseRootFromSymbol extracts just the root note (C, C#, Db, etc.)
func parseRootFromSymbol(symbol string) string {
	if len(symbol) == 0 {
		return "C"
	}

	root := string(symbol[0])
	if len(symbol) > 1 && (symbol[1] == '#' || symbol[1] == 'b') {
		root += string(symbol[1])
	}
	return root
}

// GetPlayableStrings returns which strings can be played for this voicing
func (v GuitarVoicing) GetPlayableStrings() []int {
	var strings []int
	for i := 0; i < 6; i++ {
		if v.Frets[i] >= 0 {
			strings = append(strings, i)
		}
	}
	return strings
}

// GetBassNote returns the MIDI note for the bass string
func (v GuitarVoicing) GetBassNote(tuning theory.Tuning, capo int) int {
	if v.BassString >= 0 && v.BassString < 6 && v.Frets[v.BassString] >= 0 {
		return GetFretNoteWithCapo(tuning, v.BassString, v.Frets[v.BassString], capo)
	}
	// Find lowest playable string
	for i := 0; i < 6; i++ {
		if v.Frets[i] >= 0 {
			return GetFretNoteWithCapo(tuning, i, v.Frets[i], capo)
		}
	}
	return -1
}

// GetNotes returns all MIDI notes in this voicing (low to high)
func (v GuitarVoicing) GetNotes(tuning theory.Tuning, capo int) []int {
	var notes []int
	for i := 0; i < 6; i++ {
		if v.Frets[i] >= 0 {
			notes = append(notes, GetFretNoteWithCapo(tuning, i, v.Frets[i], capo))
		}
	}
	return notes
}

// GetNoteForString returns the MIDI note for a specific string, or -1 if muted
func (v GuitarVoicing) GetNoteForString(stringNum int, tuning theory.Tuning, capo int) int {
	if stringNum < 0 || stringNum >= 6 || v.Frets[stringNum] < 0 {
		return -1
	}
	return GetFretNoteWithCapo(tuning, stringNum, v.Frets[stringNum], capo)
}
