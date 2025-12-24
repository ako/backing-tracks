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
	voicings       map[string][]ChordVoicing            // Standard tuning voicings
	tuningVoicings map[string]map[string][]ChordVoicing // Tuning-specific voicings [tuning][chord]
}

// NewChordChart creates a new chord chart with common voicings
func NewChordChart() *ChordChart {
	cc := &ChordChart{
		voicings:       make(map[string][]ChordVoicing),
		tuningVoicings: make(map[string]map[string][]ChordVoicing),
	}
	cc.loadVoicings()
	cc.loadTuningVoicings()
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
	cc.voicings["Bsus2"] = []ChordVoicing{
		{Name: "Bsus2", Frets: [6]int{-1, 2, 4, 4, 2, 2}, BaseFret: 2},
	}
	cc.voicings["Csus2"] = []ChordVoicing{
		{Name: "Csus2", Frets: [6]int{-1, 3, 0, 0, 3, 3}, BaseFret: 0},
		{Name: "Csus2 (bar)", Frets: [6]int{-1, 3, 5, 5, 3, 3}, BaseFret: 3},
	}
	cc.voicings["Dsus2"] = []ChordVoicing{
		{Name: "Dsus2", Frets: [6]int{-1, -1, 0, 2, 3, 0}, BaseFret: 0},
	}
	cc.voicings["Esus2"] = []ChordVoicing{
		{Name: "Esus2", Frets: [6]int{0, 2, 4, 4, 0, 0}, BaseFret: 0},
	}
	cc.voicings["Fsus2"] = []ChordVoicing{
		{Name: "Fsus2", Frets: [6]int{-1, -1, 3, 0, 1, 1}, BaseFret: 0},
		{Name: "Fsus2 (bar)", Frets: [6]int{1, 3, 3, 0, 1, 1}, BaseFret: 1},
	}
	cc.voicings["Gsus2"] = []ChordVoicing{
		{Name: "Gsus2", Frets: [6]int{3, 0, 0, 0, 3, 3}, BaseFret: 0},
		{Name: "Gsus2 (bar)", Frets: [6]int{3, 5, 5, 2, 3, 3}, BaseFret: 3},
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

// loadTuningVoicings populates tuning-specific chord voicings
func (cc *ChordChart) loadTuningVoicings() {
	// Drop D tuning voicings
	// Low string is now D, allowing for different fingerings
	dropD := make(map[string][]ChordVoicing)

	// Power chords are easier in drop D (one finger barre)
	dropD["D5"] = []ChordVoicing{
		{Name: "D5", Frets: [6]int{0, 0, 0, -1, -1, -1}, BaseFret: 0},
	}
	dropD["E5"] = []ChordVoicing{
		{Name: "E5", Frets: [6]int{2, 2, 2, -1, -1, -1}, BaseFret: 2},
	}
	dropD["F5"] = []ChordVoicing{
		{Name: "F5", Frets: [6]int{3, 3, 3, -1, -1, -1}, BaseFret: 3},
	}
	dropD["G5"] = []ChordVoicing{
		{Name: "G5", Frets: [6]int{5, 5, 5, -1, -1, -1}, BaseFret: 5},
	}
	dropD["A5"] = []ChordVoicing{
		{Name: "A5", Frets: [6]int{7, 7, 7, -1, -1, -1}, BaseFret: 7},
	}
	dropD["B5"] = []ChordVoicing{
		{Name: "B5", Frets: [6]int{9, 9, 9, -1, -1, -1}, BaseFret: 9},
	}
	dropD["C5"] = []ChordVoicing{
		{Name: "C5", Frets: [6]int{10, 10, 10, -1, -1, -1}, BaseFret: 10},
	}

	// D-based chords with open low D
	dropD["D"] = []ChordVoicing{
		{Name: "D", Frets: [6]int{0, 0, 0, 2, 3, 2}, BaseFret: 0},
	}
	dropD["Dm"] = []ChordVoicing{
		{Name: "Dm", Frets: [6]int{0, 0, 0, 2, 3, 1}, BaseFret: 0},
	}
	dropD["D7"] = []ChordVoicing{
		{Name: "D7", Frets: [6]int{0, 0, 0, 2, 1, 2}, BaseFret: 0},
	}
	dropD["Dmaj7"] = []ChordVoicing{
		{Name: "Dmaj7", Frets: [6]int{0, 0, 0, 2, 2, 2}, BaseFret: 0},
	}
	dropD["Dm7"] = []ChordVoicing{
		{Name: "Dm7", Frets: [6]int{0, 0, 0, 2, 1, 1}, BaseFret: 0},
	}
	dropD["Dsus2"] = []ChordVoicing{
		{Name: "Dsus2", Frets: [6]int{0, 0, 0, 2, 3, 0}, BaseFret: 0},
	}
	dropD["Dsus4"] = []ChordVoicing{
		{Name: "Dsus4", Frets: [6]int{0, 0, 0, 2, 3, 3}, BaseFret: 0},
	}

	// Other common chords in drop D
	dropD["G"] = []ChordVoicing{
		{Name: "G", Frets: [6]int{5, 5, 5, 4, 3, 3}, BaseFret: 3},
		{Name: "G (open)", Frets: [6]int{5, 2, 0, 0, 0, 3}, BaseFret: 0},
	}
	dropD["A"] = []ChordVoicing{
		{Name: "A", Frets: [6]int{7, 7, 7, 6, 5, 5}, BaseFret: 5},
		{Name: "A (open)", Frets: [6]int{-1, 0, 2, 2, 2, 0}, BaseFret: 0},
	}
	dropD["Am"] = []ChordVoicing{
		{Name: "Am", Frets: [6]int{7, 7, 7, 5, 5, 5}, BaseFret: 5},
		{Name: "Am (open)", Frets: [6]int{-1, 0, 2, 2, 1, 0}, BaseFret: 0},
	}
	dropD["E"] = []ChordVoicing{
		{Name: "E", Frets: [6]int{2, 2, 2, 1, 0, 0}, BaseFret: 0},
	}
	dropD["Em"] = []ChordVoicing{
		{Name: "Em", Frets: [6]int{2, 2, 2, 0, 0, 0}, BaseFret: 0},
	}
	dropD["F"] = []ChordVoicing{
		{Name: "F", Frets: [6]int{3, 3, 3, 2, 1, 1}, BaseFret: 1},
	}
	dropD["Bm"] = []ChordVoicing{
		{Name: "Bm", Frets: [6]int{9, 9, 9, 7, 7, 7}, BaseFret: 7},
	}
	dropD["C"] = []ChordVoicing{
		{Name: "C", Frets: [6]int{10, 10, 10, 9, 8, 8}, BaseFret: 8},
		{Name: "C (open)", Frets: [6]int{-1, 3, 2, 0, 1, 0}, BaseFret: 0},
	}

	// Sus2 chords common in rock (like Everlong)
	dropD["Bsus2"] = []ChordVoicing{
		{Name: "Bsus2", Frets: [6]int{9, 9, 9, 8, 7, 7}, BaseFret: 7},
		{Name: "Bsus2 (alt)", Frets: [6]int{-1, 2, 4, 4, 2, 2}, BaseFret: 2},
	}
	dropD["Gsus2"] = []ChordVoicing{
		{Name: "Gsus2", Frets: [6]int{5, 5, 5, 4, 3, 3}, BaseFret: 3},
		{Name: "Gsus2 (open)", Frets: [6]int{5, 0, 0, 0, 3, 3}, BaseFret: 0},
	}
	dropD["Asus2"] = []ChordVoicing{
		{Name: "Asus2", Frets: [6]int{7, 7, 7, 6, 5, 5}, BaseFret: 5},
		{Name: "Asus2 (open)", Frets: [6]int{-1, 0, 2, 2, 0, 0}, BaseFret: 0},
	}
	dropD["Esus2"] = []ChordVoicing{
		{Name: "Esus2", Frets: [6]int{2, 2, 4, 4, 0, 0}, BaseFret: 0},
	}

	cc.tuningVoicings["drop_d"] = dropD

	// Open G tuning voicings (Keith Richards style)
	openG := make(map[string][]ChordVoicing)
	openG["G"] = []ChordVoicing{
		{Name: "G", Frets: [6]int{0, 0, 0, 0, 0, 0}, BaseFret: 0},
	}
	openG["A"] = []ChordVoicing{
		{Name: "A", Frets: [6]int{2, 2, 2, 2, 2, 2}, BaseFret: 2},
	}
	openG["C"] = []ChordVoicing{
		{Name: "C", Frets: [6]int{5, 5, 5, 5, 5, 5}, BaseFret: 5},
	}
	openG["D"] = []ChordVoicing{
		{Name: "D", Frets: [6]int{7, 7, 7, 7, 7, 7}, BaseFret: 7},
	}
	cc.tuningVoicings["open_g"] = openG

	// Open D tuning voicings
	openD := make(map[string][]ChordVoicing)
	openD["D"] = []ChordVoicing{
		{Name: "D", Frets: [6]int{0, 0, 0, 0, 0, 0}, BaseFret: 0},
	}
	openD["E"] = []ChordVoicing{
		{Name: "E", Frets: [6]int{2, 2, 2, 2, 2, 2}, BaseFret: 2},
	}
	openD["G"] = []ChordVoicing{
		{Name: "G", Frets: [6]int{5, 5, 5, 5, 5, 5}, BaseFret: 5},
	}
	openD["A"] = []ChordVoicing{
		{Name: "A", Frets: [6]int{7, 7, 7, 7, 7, 7}, BaseFret: 7},
	}
	cc.tuningVoicings["open_d"] = openD

	// Open E tuning voicings
	openE := make(map[string][]ChordVoicing)
	openE["E"] = []ChordVoicing{
		{Name: "E", Frets: [6]int{0, 0, 0, 0, 0, 0}, BaseFret: 0},
	}
	openE["A"] = []ChordVoicing{
		{Name: "A", Frets: [6]int{5, 5, 5, 5, 5, 5}, BaseFret: 5},
	}
	openE["B"] = []ChordVoicing{
		{Name: "B", Frets: [6]int{7, 7, 7, 7, 7, 7}, BaseFret: 7},
	}
	cc.tuningVoicings["open_e"] = openE

	// DADGAD tuning voicings
	dadgad := make(map[string][]ChordVoicing)
	dadgad["D"] = []ChordVoicing{
		{Name: "D", Frets: [6]int{0, 0, 0, 0, 0, 0}, BaseFret: 0},
	}
	dadgad["Dsus4"] = []ChordVoicing{
		{Name: "Dsus4", Frets: [6]int{0, 0, 0, 0, 0, 0}, BaseFret: 0},
	}
	dadgad["G"] = []ChordVoicing{
		{Name: "G", Frets: [6]int{5, 5, 5, 5, 5, 5}, BaseFret: 5},
	}
	dadgad["A"] = []ChordVoicing{
		{Name: "A", Frets: [6]int{7, 7, 7, 7, 7, 7}, BaseFret: 7},
	}
	cc.tuningVoicings["dadgad"] = dadgad
}

// GetVoicings returns all voicings for a chord symbol (standard tuning)
func (cc *ChordChart) GetVoicings(symbol string) []ChordVoicing {
	return cc.GetVoicingsForTuning(symbol, "standard")
}

// GetVoicingsForTuning returns voicings for a chord in a specific tuning
func (cc *ChordChart) GetVoicingsForTuning(symbol, tuning string) []ChordVoicing {
	// Check for tuning-specific voicings first
	if tuning != "" && tuning != "standard" {
		if tuningChords, ok := cc.tuningVoicings[tuning]; ok {
			if voicings, ok := tuningChords[symbol]; ok {
				return voicings
			}
			// Try normalized
			normalized := normalizeChordSymbol(symbol)
			if voicings, ok := tuningChords[normalized]; ok {
				return voicings
			}
		}
	}

	// Fall back to standard voicings
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
