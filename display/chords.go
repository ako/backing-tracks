package display

import (
	"fmt"
	"strings"
)

// ChordVoicing represents finger positions for a chord
// Format: [E, A, D, G, B, e] where -1 = muted, 0 = open, 1+ = fret
type ChordVoicing struct {
	Name     string
	Frets    [6]int // -1 = x (muted), 0 = open, 1+ = fret number
	BaseFret int    // For barre chords, the starting fret (0 for open position)
	Fingers  string // Optional finger positions
}

// ChordChart manages chord diagram display
type ChordChart struct {
	voicings map[string][]ChordVoicing // Multiple voicings per chord
}

// NewChordChart creates a new chord chart with common voicings
func NewChordChart() *ChordChart {
	cc := &ChordChart{
		voicings: make(map[string][]ChordVoicing),
	}
	cc.loadVoicings()
	return cc
}

// loadVoicings populates common chord voicings
func (cc *ChordChart) loadVoicings() {
	// Major chords
	cc.voicings["A"] = []ChordVoicing{
		{Name: "A", Frets: [6]int{-1, 0, 2, 2, 2, 0}, BaseFret: 0},
		{Name: "A (bar)", Frets: [6]int{5, 7, 7, 6, 5, 5}, BaseFret: 5},
	}
	cc.voicings["B"] = []ChordVoicing{
		{Name: "B", Frets: [6]int{-1, 2, 4, 4, 4, 2}, BaseFret: 2},
		{Name: "B (bar)", Frets: [6]int{7, 9, 9, 8, 7, 7}, BaseFret: 7},
	}
	cc.voicings["C"] = []ChordVoicing{
		{Name: "C", Frets: [6]int{-1, 3, 2, 0, 1, 0}, BaseFret: 0},
		{Name: "C (bar)", Frets: [6]int{8, 10, 10, 9, 8, 8}, BaseFret: 8},
	}
	cc.voicings["D"] = []ChordVoicing{
		{Name: "D", Frets: [6]int{-1, -1, 0, 2, 3, 2}, BaseFret: 0},
		{Name: "D (bar)", Frets: [6]int{-1, 5, 7, 7, 7, 5}, BaseFret: 5},
	}
	cc.voicings["E"] = []ChordVoicing{
		{Name: "E", Frets: [6]int{0, 2, 2, 1, 0, 0}, BaseFret: 0},
		{Name: "E (7th)", Frets: [6]int{-1, 7, 9, 9, 9, 7}, BaseFret: 7},
	}
	cc.voicings["F"] = []ChordVoicing{
		{Name: "F", Frets: [6]int{1, 3, 3, 2, 1, 1}, BaseFret: 1},
		{Name: "F (easy)", Frets: [6]int{-1, -1, 3, 2, 1, 1}, BaseFret: 1},
	}
	cc.voicings["G"] = []ChordVoicing{
		{Name: "G", Frets: [6]int{3, 2, 0, 0, 0, 3}, BaseFret: 0},
		{Name: "G (bar)", Frets: [6]int{3, 5, 5, 4, 3, 3}, BaseFret: 3},
	}

	// Minor chords
	cc.voicings["Am"] = []ChordVoicing{
		{Name: "Am", Frets: [6]int{-1, 0, 2, 2, 1, 0}, BaseFret: 0},
		{Name: "Am (bar)", Frets: [6]int{5, 7, 7, 5, 5, 5}, BaseFret: 5},
	}
	cc.voicings["Bm"] = []ChordVoicing{
		{Name: "Bm", Frets: [6]int{-1, 2, 4, 4, 3, 2}, BaseFret: 2},
		{Name: "Bm (bar)", Frets: [6]int{7, 9, 9, 7, 7, 7}, BaseFret: 7},
	}
	cc.voicings["Cm"] = []ChordVoicing{
		{Name: "Cm", Frets: [6]int{-1, 3, 5, 5, 4, 3}, BaseFret: 3},
		{Name: "Cm (8th)", Frets: [6]int{8, 10, 10, 8, 8, 8}, BaseFret: 8},
	}
	cc.voicings["Dm"] = []ChordVoicing{
		{Name: "Dm", Frets: [6]int{-1, -1, 0, 2, 3, 1}, BaseFret: 0},
		{Name: "Dm (bar)", Frets: [6]int{-1, 5, 7, 7, 6, 5}, BaseFret: 5},
	}
	cc.voicings["Em"] = []ChordVoicing{
		{Name: "Em", Frets: [6]int{0, 2, 2, 0, 0, 0}, BaseFret: 0},
		{Name: "Em (7th)", Frets: [6]int{-1, 7, 9, 9, 8, 7}, BaseFret: 7},
	}
	cc.voicings["Fm"] = []ChordVoicing{
		{Name: "Fm", Frets: [6]int{1, 3, 3, 1, 1, 1}, BaseFret: 1},
	}
	cc.voicings["Gm"] = []ChordVoicing{
		{Name: "Gm", Frets: [6]int{3, 5, 5, 3, 3, 3}, BaseFret: 3},
	}

	// Dominant 7th chords
	cc.voicings["A7"] = []ChordVoicing{
		{Name: "A7", Frets: [6]int{-1, 0, 2, 0, 2, 0}, BaseFret: 0},
		{Name: "A7 (bar)", Frets: [6]int{5, 7, 5, 6, 5, 5}, BaseFret: 5},
	}
	cc.voicings["B7"] = []ChordVoicing{
		{Name: "B7", Frets: [6]int{-1, 2, 1, 2, 0, 2}, BaseFret: 0},
		{Name: "B7 (bar)", Frets: [6]int{7, 9, 7, 8, 7, 7}, BaseFret: 7},
	}
	cc.voicings["C7"] = []ChordVoicing{
		{Name: "C7", Frets: [6]int{-1, 3, 2, 3, 1, 0}, BaseFret: 0},
		{Name: "C7 (bar)", Frets: [6]int{8, 10, 8, 9, 8, 8}, BaseFret: 8},
	}
	cc.voicings["D7"] = []ChordVoicing{
		{Name: "D7", Frets: [6]int{-1, -1, 0, 2, 1, 2}, BaseFret: 0},
		{Name: "D7 (bar)", Frets: [6]int{-1, 5, 7, 5, 7, 5}, BaseFret: 5},
	}
	cc.voicings["E7"] = []ChordVoicing{
		{Name: "E7", Frets: [6]int{0, 2, 0, 1, 0, 0}, BaseFret: 0},
		{Name: "E7 (bar)", Frets: [6]int{-1, 7, 9, 7, 9, 7}, BaseFret: 7},
	}
	cc.voicings["F7"] = []ChordVoicing{
		{Name: "F7", Frets: [6]int{1, 3, 1, 2, 1, 1}, BaseFret: 1},
	}
	cc.voicings["G7"] = []ChordVoicing{
		{Name: "G7", Frets: [6]int{3, 2, 0, 0, 0, 1}, BaseFret: 0},
		{Name: "G7 (bar)", Frets: [6]int{3, 5, 3, 4, 3, 3}, BaseFret: 3},
	}

	// Minor 7th chords
	cc.voicings["Am7"] = []ChordVoicing{
		{Name: "Am7", Frets: [6]int{-1, 0, 2, 0, 1, 0}, BaseFret: 0},
		{Name: "Am7 (bar)", Frets: [6]int{5, 7, 5, 5, 5, 5}, BaseFret: 5},
	}
	cc.voicings["Bm7"] = []ChordVoicing{
		{Name: "Bm7", Frets: [6]int{-1, 2, 4, 2, 3, 2}, BaseFret: 2},
	}
	cc.voicings["Cm7"] = []ChordVoicing{
		{Name: "Cm7", Frets: [6]int{-1, 3, 5, 3, 4, 3}, BaseFret: 3},
	}
	cc.voicings["Dm7"] = []ChordVoicing{
		{Name: "Dm7", Frets: [6]int{-1, -1, 0, 2, 1, 1}, BaseFret: 0},
		{Name: "Dm7 (bar)", Frets: [6]int{-1, 5, 7, 5, 6, 5}, BaseFret: 5},
	}
	cc.voicings["Em7"] = []ChordVoicing{
		{Name: "Em7", Frets: [6]int{0, 2, 0, 0, 0, 0}, BaseFret: 0},
		{Name: "Em7 (bar)", Frets: [6]int{-1, 7, 9, 7, 8, 7}, BaseFret: 7},
	}
	cc.voicings["Fm7"] = []ChordVoicing{
		{Name: "Fm7", Frets: [6]int{1, 3, 1, 1, 1, 1}, BaseFret: 1},
	}
	cc.voicings["Gm7"] = []ChordVoicing{
		{Name: "Gm7", Frets: [6]int{3, 5, 3, 3, 3, 3}, BaseFret: 3},
	}

	// Major 7th chords
	cc.voicings["Amaj7"] = []ChordVoicing{
		{Name: "Amaj7", Frets: [6]int{-1, 0, 2, 1, 2, 0}, BaseFret: 0},
	}
	cc.voicings["Cmaj7"] = []ChordVoicing{
		{Name: "Cmaj7", Frets: [6]int{-1, 3, 2, 0, 0, 0}, BaseFret: 0},
	}
	cc.voicings["Dmaj7"] = []ChordVoicing{
		{Name: "Dmaj7", Frets: [6]int{-1, -1, 0, 2, 2, 2}, BaseFret: 0},
	}
	cc.voicings["Emaj7"] = []ChordVoicing{
		{Name: "Emaj7", Frets: [6]int{0, 2, 1, 1, 0, 0}, BaseFret: 0},
	}
	cc.voicings["Fmaj7"] = []ChordVoicing{
		{Name: "Fmaj7", Frets: [6]int{-1, -1, 3, 2, 1, 0}, BaseFret: 0},
		{Name: "Fmaj7 (bar)", Frets: [6]int{1, 3, 2, 2, 1, 1}, BaseFret: 1},
	}
	cc.voicings["Gmaj7"] = []ChordVoicing{
		{Name: "Gmaj7", Frets: [6]int{3, 2, 0, 0, 0, 2}, BaseFret: 0},
	}

	// Suspended chords
	cc.voicings["Asus4"] = []ChordVoicing{
		{Name: "Asus4", Frets: [6]int{-1, 0, 2, 2, 3, 0}, BaseFret: 0},
	}
	cc.voicings["Dsus4"] = []ChordVoicing{
		{Name: "Dsus4", Frets: [6]int{-1, -1, 0, 2, 3, 3}, BaseFret: 0},
	}
	cc.voicings["Esus4"] = []ChordVoicing{
		{Name: "Esus4", Frets: [6]int{0, 2, 2, 2, 0, 0}, BaseFret: 0},
	}
	cc.voicings["Asus2"] = []ChordVoicing{
		{Name: "Asus2", Frets: [6]int{-1, 0, 2, 2, 0, 0}, BaseFret: 0},
	}
	cc.voicings["Dsus2"] = []ChordVoicing{
		{Name: "Dsus2", Frets: [6]int{-1, -1, 0, 2, 3, 0}, BaseFret: 0},
	}

	// Add aliases for flat/sharp variants
	cc.voicings["Bb"] = cc.voicings["A#"]
	cc.voicings["Db"] = cc.voicings["C#"]
	cc.voicings["Eb"] = cc.voicings["D#"]
	cc.voicings["Gb"] = cc.voicings["F#"]
	cc.voicings["Ab"] = cc.voicings["G#"]

	// Sharp variants using barre shapes
	cc.voicings["A#"] = []ChordVoicing{
		{Name: "A#/Bb", Frets: [6]int{-1, 1, 3, 3, 3, 1}, BaseFret: 1},
	}
	cc.voicings["C#"] = []ChordVoicing{
		{Name: "C#/Db", Frets: [6]int{-1, 4, 6, 6, 6, 4}, BaseFret: 4},
	}
	cc.voicings["D#"] = []ChordVoicing{
		{Name: "D#/Eb", Frets: [6]int{-1, 6, 8, 8, 8, 6}, BaseFret: 6},
	}
	cc.voicings["F#"] = []ChordVoicing{
		{Name: "F#/Gb", Frets: [6]int{2, 4, 4, 3, 2, 2}, BaseFret: 2},
	}
	cc.voicings["G#"] = []ChordVoicing{
		{Name: "G#/Ab", Frets: [6]int{4, 6, 6, 5, 4, 4}, BaseFret: 4},
	}

	// Update flat aliases
	cc.voicings["Bb"] = cc.voicings["A#"]
	cc.voicings["Db"] = cc.voicings["C#"]
	cc.voicings["Eb"] = cc.voicings["D#"]
	cc.voicings["Gb"] = cc.voicings["F#"]
	cc.voicings["Ab"] = cc.voicings["G#"]
}

// GetVoicings returns all voicings for a chord symbol
func (cc *ChordChart) GetVoicings(symbol string) []ChordVoicing {
	// Try exact match first
	if voicings, ok := cc.voicings[symbol]; ok {
		return voicings
	}

	// Try normalized version (handle variations like Amin, Ami -> Am)
	normalized := normalizeChordSymbol(symbol)
	if voicings, ok := cc.voicings[normalized]; ok {
		return voicings
	}

	return nil
}

// normalizeChordSymbol converts chord variations to standard form
func normalizeChordSymbol(symbol string) string {
	// Replace common variations
	s := symbol
	s = strings.Replace(s, "min", "m", 1)
	s = strings.Replace(s, "mi", "m", 1)
	s = strings.Replace(s, "maj", "maj", 1)
	s = strings.Replace(s, "M7", "maj7", 1)
	return s
}

// RenderHorizontal renders a chord diagram horizontally
// Returns multiple lines for display
func (cc *ChordChart) RenderHorizontal(symbol string) []string {
	voicings := cc.GetVoicings(symbol)
	if len(voicings) == 0 {
		return []string{fmt.Sprintf(" %s: [no chart]", symbol)}
	}

	lines := []string{}

	// Show first voicing (main position)
	v := voicings[0]
	lines = append(lines, cc.RenderSingleChord(v)...)

	// If there's a second voicing, show it too
	if len(voicings) > 1 {
		lines = append(lines, "") // spacer
		lines = append(lines, cc.RenderSingleChord(voicings[1])...)
	}

	return lines
}

// RenderSingleChord renders one chord voicing horizontally
func (cc *ChordChart) RenderSingleChord(v ChordVoicing) []string {
	lines := []string{}

	// Chord name and tab notation
	tabStr := ""
	for i := 0; i < 6; i++ {
		if v.Frets[i] == -1 {
			tabStr += "x"
		} else {
			tabStr += fmt.Sprintf("%d", v.Frets[i])
		}
	}
	lines = append(lines, fmt.Sprintf(" \033[1m%s\033[0m [%s]", v.Name, tabStr))

	// Find fret range to display
	minFret := 99
	maxFret := 0
	for _, f := range v.Frets {
		if f > 0 {
			if f < minFret {
				minFret = f
			}
			if f > maxFret {
				maxFret = f
			}
		}
	}

	// Determine display range (show 4 frets)
	startFret := 1
	if v.BaseFret > 0 {
		startFret = v.BaseFret
	} else if minFret > 3 {
		startFret = minFret - 1
	}
	endFret := startFret + 3

	// Open/muted string indicators (above the nut)
	indicatorLine := " "
	for str := 0; str < 6; str++ {
		f := v.Frets[str]
		if f == -1 {
			indicatorLine += "x  "
		} else if f == 0 {
			indicatorLine += "○  "
		} else {
			indicatorLine += "   "
		}
	}
	lines = append(lines, indicatorLine)

	// Nut or fret number indicator
	if startFret == 1 {
		lines = append(lines, " ══════════════════")
	} else {
		lines = append(lines, fmt.Sprintf(" %dfr─────────────", startFret))
	}

	// Draw frets
	for fret := startFret; fret <= endFret; fret++ {
		line := " "
		for str := 0; str < 6; str++ {
			f := v.Frets[str]
			if f == fret {
				line += "●  "
			} else {
				line += "│  "
			}
		}
		lines = append(lines, line)
	}

	return lines
}

// RenderCompact renders a very compact single-line chord indicator
func (cc *ChordChart) RenderCompact(symbol string) string {
	voicings := cc.GetVoicings(symbol)
	if len(voicings) == 0 {
		return fmt.Sprintf("%s: ?", symbol)
	}

	v := voicings[0]
	tabStr := ""
	for i := 0; i < 6; i++ {
		if v.Frets[i] == -1 {
			tabStr += "x"
		} else {
			tabStr += fmt.Sprintf("%d", v.Frets[i])
		}
	}
	return fmt.Sprintf("%s [%s]", v.Name, tabStr)
}
