package midi

import (
	"backing-tracks/theory"
)

// PatternType defines the type of fingerstyle pattern
type PatternType string

const (
	PatternTravis     PatternType = "travis"
	PatternArpeggio   PatternType = "arpeggio"
	PatternFolk       PatternType = "folk"
	PatternClassical  PatternType = "classical"
	PatternBossaNova  PatternType = "bossa"
	PatternBallad     PatternType = "ballad"
	PatternWaltz      PatternType = "waltz"
	PatternFingerpick PatternType = "fingerpick"
)

// PatternNote represents a single note in a picking pattern
type PatternNote struct {
	String   int     // 0-5, where 0=low E, 5=high e
	Beat     float64 // Position within bar (1.0, 1.5, 2.0, etc.)
	Duration float64 // Duration in beats
	IsBass   bool    // Is this a bass note (thumb)
	Finger   string  // p=thumb, i=index, m=middle, a=ring
	Velocity int     // MIDI velocity (1-127)
}

// FingerstylePattern defines a complete picking pattern for one bar
type FingerstylePattern struct {
	Name         string
	Type         PatternType
	TimeSignature string // "4/4", "3/4", "6/8"
	Notes        []PatternNote
	Description  string
}

// PatternLibrary contains all available fingerstyle patterns
var PatternLibrary = map[PatternType][]FingerstylePattern{
	PatternTravis: {
		{
			Name:          "Travis Basic",
			Type:          PatternTravis,
			TimeSignature: "4/4",
			Description:   "Alternating bass with melody on off-beats",
			Notes: []PatternNote{
				// Beat 1: Bass on low string
				{String: 0, Beat: 1.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 90},
				// Beat 1.5: Melody/chord
				{String: 2, Beat: 1.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 4, Beat: 1.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				// Beat 2: Alternating bass
				{String: 1, Beat: 2.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				// Beat 2.5: Melody/chord
				{String: 3, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 5, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				// Beat 3: Bass
				{String: 0, Beat: 3.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 90},
				// Beat 3.5: Melody
				{String: 2, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 4, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				// Beat 4: Alternating bass
				{String: 1, Beat: 4.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				// Beat 4.5: Melody
				{String: 3, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 5, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
			},
		},
		{
			Name:          "Travis Simple",
			Type:          PatternTravis,
			TimeSignature: "4/4",
			Description:   "Simplified Travis picking with single melody notes",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 90},
				{String: 4, Beat: 1.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 1, Beat: 2.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				{String: 5, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 0, Beat: 3.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 90},
				{String: 4, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 1, Beat: 4.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				{String: 5, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
			},
		},
	},
	PatternArpeggio: {
		{
			Name:          "Arpeggio Up",
			Type:          PatternArpeggio,
			TimeSignature: "4/4",
			Description:   "Rolling arpeggio from bass to treble",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 90},
				{String: 2, Beat: 1.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 75},
				{String: 3, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 4, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 70},
				{String: 5, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 75},
				{String: 4, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 3, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 2, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
			},
		},
		{
			Name:          "Arpeggio PIMA",
			Type:          PatternArpeggio,
			TimeSignature: "4/4",
			Description:   "Classical PIMA arpeggio pattern",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 90},
				{String: 2, Beat: 1.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 75},
				{String: 3, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 4, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 70},
				{String: 1, Beat: 3.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				{String: 2, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 75},
				{String: 3, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 4, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 70},
			},
		},
	},
	PatternFolk: {
		{
			Name:          "Folk Ballad",
			Type:          PatternFolk,
			TimeSignature: "4/4",
			Description:   "Simple folk picking pattern",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 90},
				{String: 3, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 4, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 5, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 75},
				{String: 4, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 3, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 4, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
			},
		},
	},
	PatternClassical: {
		{
			Name:          "Classical Tremolo",
			Type:          PatternClassical,
			TimeSignature: "4/4",
			Description:   "p-a-m-i tremolo pattern",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 0.25, IsBass: true, Finger: "p", Velocity: 90},
				{String: 5, Beat: 1.25, Duration: 0.25, IsBass: false, Finger: "a", Velocity: 75},
				{String: 5, Beat: 1.5, Duration: 0.25, IsBass: false, Finger: "m", Velocity: 70},
				{String: 5, Beat: 1.75, Duration: 0.25, IsBass: false, Finger: "i", Velocity: 70},
				{String: 1, Beat: 2.0, Duration: 0.25, IsBass: true, Finger: "p", Velocity: 85},
				{String: 5, Beat: 2.25, Duration: 0.25, IsBass: false, Finger: "a", Velocity: 75},
				{String: 5, Beat: 2.5, Duration: 0.25, IsBass: false, Finger: "m", Velocity: 70},
				{String: 5, Beat: 2.75, Duration: 0.25, IsBass: false, Finger: "i", Velocity: 70},
				{String: 0, Beat: 3.0, Duration: 0.25, IsBass: true, Finger: "p", Velocity: 90},
				{String: 5, Beat: 3.25, Duration: 0.25, IsBass: false, Finger: "a", Velocity: 75},
				{String: 5, Beat: 3.5, Duration: 0.25, IsBass: false, Finger: "m", Velocity: 70},
				{String: 5, Beat: 3.75, Duration: 0.25, IsBass: false, Finger: "i", Velocity: 70},
				{String: 1, Beat: 4.0, Duration: 0.25, IsBass: true, Finger: "p", Velocity: 85},
				{String: 5, Beat: 4.25, Duration: 0.25, IsBass: false, Finger: "a", Velocity: 75},
				{String: 5, Beat: 4.5, Duration: 0.25, IsBass: false, Finger: "m", Velocity: 70},
				{String: 5, Beat: 4.75, Duration: 0.25, IsBass: false, Finger: "i", Velocity: 70},
			},
		},
	},
	PatternBossaNova: {
		{
			Name:          "Bossa Nova Basic",
			Type:          PatternBossaNova,
			TimeSignature: "4/4",
			Description:   "Brazilian bossa nova syncopated pattern",
			Notes: []PatternNote{
				{String: 1, Beat: 1.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 90},
				{String: 3, Beat: 1.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 4, Beat: 1.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 3, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 65},
				{String: 4, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
				{String: 0, Beat: 2.5, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				{String: 3, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 4, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 1, Beat: 4.0, Duration: 0.5, IsBass: true, Finger: "p", Velocity: 85},
				{String: 3, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 65},
				{String: 4, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
			},
		},
	},
	PatternBallad: {
		{
			Name:          "Ballad 6/8",
			Type:          PatternBallad,
			TimeSignature: "6/8",
			Description:   "Gentle 6/8 ballad pattern",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 90},
				{String: 2, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 3, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
				{String: 4, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 70},
				{String: 3, Beat: 5.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
				{String: 2, Beat: 6.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 60},
			},
		},
		{
			Name:          "Ballad 4/4",
			Type:          PatternBallad,
			TimeSignature: "4/4",
			Description:   "Slow ballad with sustained bass",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 2.0, IsBass: true, Finger: "p", Velocity: 90},
				{String: 2, Beat: 1.5, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 65},
				{String: 3, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 4, Beat: 2.5, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 65},
				{String: 1, Beat: 3.0, Duration: 2.0, IsBass: true, Finger: "p", Velocity: 85},
				{String: 3, Beat: 3.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
				{String: 4, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 70},
				{String: 5, Beat: 4.5, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
			},
		},
	},
	PatternWaltz: {
		{
			Name:          "Waltz 3/4",
			Type:          PatternWaltz,
			TimeSignature: "3/4",
			Description:   "Classic waltz bass-chord-chord pattern",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 95},
				{String: 2, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
				{String: 3, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 4, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 70},
				{String: 2, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 65},
				{String: 3, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 65},
				{String: 4, Beat: 3.0, Duration: 0.5, IsBass: false, Finger: "a", Velocity: 65},
			},
		},
	},
	PatternFingerpick: {
		{
			Name:          "Simple Fingerpick",
			Type:          PatternFingerpick,
			TimeSignature: "4/4",
			Description:   "Basic fingerpicking for beginners",
			Notes: []PatternNote{
				{String: 0, Beat: 1.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 90},
				{String: 4, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
				{String: 1, Beat: 3.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 85},
				{String: 5, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
			},
		},
	},
}

// GetPattern returns the default pattern for a given type and time signature
func GetPattern(patternType PatternType, timeSignature string) FingerstylePattern {
	patterns, ok := PatternLibrary[patternType]
	if !ok {
		// Fall back to simple fingerpick
		patterns = PatternLibrary[PatternFingerpick]
	}

	// Find pattern matching time signature
	for _, p := range patterns {
		if p.TimeSignature == timeSignature {
			return p
		}
	}

	// Return first pattern if no time signature match
	if len(patterns) > 0 {
		return patterns[0]
	}

	// Ultimate fallback
	return FingerstylePattern{
		Name:          "Default",
		Type:          PatternFingerpick,
		TimeSignature: "4/4",
		Notes: []PatternNote{
			{String: 0, Beat: 1.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 90},
			{String: 4, Beat: 2.0, Duration: 0.5, IsBass: false, Finger: "m", Velocity: 70},
			{String: 1, Beat: 3.0, Duration: 1.0, IsBass: true, Finger: "p", Velocity: 85},
			{String: 5, Beat: 4.0, Duration: 0.5, IsBass: false, Finger: "i", Velocity: 70},
		},
	}
}

// GetPatternForStyle returns an appropriate pattern based on music style
func GetPatternForStyle(style string, timeSignature string) FingerstylePattern {
	var patternType PatternType

	switch style {
	case "blues", "rock", "country":
		patternType = PatternTravis
	case "folk", "acoustic":
		patternType = PatternFolk
	case "classical", "spanish", "flamenco":
		patternType = PatternClassical
	case "bossa", "bossa_nova", "latin", "jazz":
		patternType = PatternBossaNova
	case "ballad", "slow", "pop_ballad":
		patternType = PatternBallad
	case "waltz":
		patternType = PatternWaltz
	default:
		patternType = PatternArpeggio
	}

	// Handle 6/8 and 3/4 time signatures
	if timeSignature == "6/8" && patternType != PatternBallad {
		patternType = PatternBallad
	}
	if timeSignature == "3/4" && patternType != PatternWaltz {
		patternType = PatternWaltz
	}

	return GetPattern(patternType, timeSignature)
}

// ApplyPatternToVoicing applies a pattern to a chord voicing, adjusting strings
// based on which strings are actually playable in the voicing
func ApplyPatternToVoicing(pattern FingerstylePattern, voicing GuitarVoicing, tuning theory.Tuning, capo int) []TabNote {
	playableStrings := voicing.GetPlayableStrings()
	if len(playableStrings) == 0 {
		return nil
	}

	var notes []TabNote

	for _, pn := range pattern.Notes {
		// Map pattern string to actual playable string
		actualString := mapPatternString(pn.String, pn.IsBass, playableStrings, voicing)
		if actualString < 0 {
			continue
		}

		fret := voicing.Frets[actualString]
		if fret < 0 {
			continue
		}

		note := TabNote{
			String:   actualString,
			Fret:     fret,
			Beat:     pn.Beat,
			Duration: pn.Duration,
			Finger:   pn.Finger,
			Velocity: pn.Velocity,
			MidiNote: GetFretNoteWithCapo(tuning, actualString, fret, capo),
		}
		notes = append(notes, note)
	}

	return notes
}

// mapPatternString maps a pattern string number to an actual playable string
func mapPatternString(patternString int, isBass bool, playableStrings []int, voicing GuitarVoicing) int {
	if len(playableStrings) == 0 {
		return -1
	}

	if isBass {
		// For bass notes, use the bass string from the voicing
		return voicing.BassString
	}

	// For treble notes, map to higher strings
	// Pattern strings 2-5 should map to the treble portion of playable strings
	trebleStrings := []int{}
	for _, s := range playableStrings {
		if s >= 2 { // Strings D, G, B, e
			trebleStrings = append(trebleStrings, s)
		}
	}

	if len(trebleStrings) == 0 {
		// No treble strings, use any playable string
		if patternString < len(playableStrings) {
			return playableStrings[patternString]
		}
		return playableStrings[len(playableStrings)-1]
	}

	// Map pattern string (2-5) to available treble strings
	idx := patternString - 2
	if idx < 0 {
		idx = 0
	}
	if idx >= len(trebleStrings) {
		idx = len(trebleStrings) - 1
	}

	return trebleStrings[idx]
}

// TabNote represents a note in tablature with all information needed for display
type TabNote struct {
	String   int     // Guitar string (0-5)
	Fret     int     // Fret number
	Beat     float64 // Position in bar
	Duration float64 // Duration in beats
	Finger   string  // Right hand finger (p, i, m, a)
	Velocity int     // MIDI velocity
	MidiNote int     // Resulting MIDI note
}

// TabBar represents one bar of tablature
type TabBar struct {
	ChordName string
	Notes     []TabNote
	BarNumber int
}
