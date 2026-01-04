package display

import (
	"fmt"
	"strings"
	"time"

	"backing-tracks/theory"
)

// updatePosition calculates current bar/beat from elapsed time
func (m *TUIModel) updatePosition() {
	// If we have a player, sync from it
	if m.player != nil {
		m.currentBar, m.currentBeat, m.currentStrum, m.paused = m.player.GetPlaybackState()
		// Update tablature position
		if m.tablature != nil {
			m.tablature.SetPosition(m.currentBar, float64(m.currentBeat)+1)
		}
		return
	}

	// Fallback: calculate from local time (display-only mode)
	elapsed := time.Since(m.startTime) - m.pausedTotal + m.seekOffset
	if elapsed < 0 {
		elapsed = 0
		// Reset seek offset to prevent going negative
		m.seekOffset = m.pausedTotal - time.Since(m.startTime)
	}
	totalBeats := int(elapsed / m.timePerBeat)
	m.currentBeat = totalBeats % 4
	m.currentBar = totalBeats / 4

	// Calculate strum position (8 or 16 strums per bar)
	strumsPerBar := 8
	if m.isSixteenthNoteStyle() {
		strumsPerBar = 16
	}
	timePerStrum := m.timePerBeat * 4 / time.Duration(strumsPerBar)
	totalStrums := int(elapsed / timePerStrum)
	m.currentStrum = totalStrums % strumsPerBar

	// Update tablature position
	if m.tablature != nil {
		m.tablature.SetPosition(m.currentBar, float64(m.currentBeat)+1)
	}
}

// getBarChordName returns the chord name(s) for a bar (with transpose applied)
func (m *TUIModel) getBarChordName(barIdx int) string {
	if barIdx >= len(m.bars) || len(m.bars[barIdx].Chords) == 0 {
		return ""
	}
	bar := m.bars[barIdx]
	if len(bar.Chords) == 1 {
		if m.transposeOffset != 0 {
			return transposeChord(bar.Chords[0].Symbol, m.transposeOffset)
		}
		return bar.Chords[0].Symbol
	}
	// Multiple chords in this bar - show all (transposed)
	var names []string
	for _, bc := range bar.Chords {
		name := bc.Symbol
		if m.transposeOffset != 0 {
			name = transposeChord(name, m.transposeOffset)
		}
		names = append(names, name)
	}
	return strings.Join(names, " → ")
}

// getCurrentChordSymbol returns the chord symbol for the current beat position (transposed)
func (m *TUIModel) getCurrentChordSymbol() string {
	if m.currentBar >= len(m.bars) || len(m.bars) == 0 {
		return ""
	}
	bar := m.bars[m.currentBar]
	if len(bar.Chords) == 0 {
		return ""
	}

	// Find the chord active at the current beat
	var symbol string
	for i := len(bar.Chords) - 1; i >= 0; i-- {
		chord := bar.Chords[i]
		if m.currentBeat >= chord.StartBeat {
			symbol = chord.Symbol
			break
		}
	}
	if symbol == "" {
		symbol = bar.Chords[0].Symbol
	}

	// Apply transpose
	if m.transposeOffset != 0 {
		return transposeChord(symbol, m.transposeOffset)
	}
	return symbol
}

// transposeChord transposes a chord symbol by the given number of semitones
func transposeChord(symbol string, semitones int) string {
	if symbol == "" {
		return ""
	}

	// Note names in order
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	flatNames := []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

	// Parse the root note
	var root string
	var remainder string
	useFlats := false

	if len(symbol) >= 2 && (symbol[1] == '#' || symbol[1] == 'b') {
		root = symbol[:2]
		remainder = symbol[2:]
		useFlats = symbol[1] == 'b'
	} else {
		root = symbol[:1]
		remainder = symbol[1:]
	}

	// Find root index
	rootUpper := strings.ToUpper(root)
	rootIdx := -1
	for i, n := range noteNames {
		if n == rootUpper || flatNames[i] == rootUpper {
			rootIdx = i
			break
		}
	}
	if rootIdx == -1 {
		return symbol // Can't transpose, return as-is
	}

	// Transpose
	newIdx := (rootIdx + semitones%12 + 12) % 12

	// Get new root name
	var newRoot string
	if useFlats {
		newRoot = flatNames[newIdx]
	} else {
		newRoot = noteNames[newIdx]
	}

	return newRoot + remainder
}

// updateTransposedScale updates the scale display when transpose changes
func (m *TUIModel) updateTransposedScale() {
	// Get the transposed key
	originalKey := m.track.Info.Key
	transposedKey := transposeChord(originalKey, m.transposeOffset)

	// Update the scale
	m.currentScale = theory.GetScaleForStyle(transposedKey, m.track.Info.Style, "")
}

// getCapoAdjustedTuning returns the tuning with capo applied
// When capo is at fret N, each string's pitch is raised by N semitones
func (m *TUIModel) getCapoAdjustedTuning() theory.Tuning {
	if m.capoPosition == 0 {
		return m.tuning
	}

	// Create new tuning with adjusted notes
	adjusted := theory.Tuning{
		Notes: make([]int, len(m.tuning.Notes)),
		Names: make([]string, len(m.tuning.Names)),
	}

	for i, note := range m.tuning.Notes {
		adjusted.Notes[i] = note + m.capoPosition
	}

	// Update string names to reflect new pitches
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	for i, note := range adjusted.Notes {
		adjusted.Names[i] = noteNames[note%12]
	}
	// Make high string lowercase if it was originally
	if len(adjusted.Names) > 0 && len(m.tuning.Names) > 0 {
		lastIdx := len(adjusted.Names) - 1
		if len(m.tuning.Names[lastIdx]) > 0 && m.tuning.Names[lastIdx][0] >= 'a' && m.tuning.Names[lastIdx][0] <= 'z' {
			adjusted.Names[lastIdx] = strings.ToLower(adjusted.Names[lastIdx])
		}
	}

	return adjusted
}

// updateTablatureConfig updates the tablature with current tuning and capo settings
func (m *TUIModel) updateTablatureConfig() {
	if m.tablature != nil {
		m.tablature.UpdateConfig(m.tuning, m.capoPosition)
		m.tablature.RegenerateTablature(m.track)
	}
}

// cycleTuning changes the tuning by the given offset (-1 for previous, +1 for next)
func (m *TUIModel) cycleTuning(offset int) {
	numTunings := len(theory.TuningNames)
	m.tuningIndex = (m.tuningIndex + offset + numTunings) % numTunings
	m.tuningName = theory.TuningNames[m.tuningIndex]
	m.tuning = theory.GetTuning(m.tuningName)

	// Update fretboard display with new tuning
	if m.fretboard != nil {
		m.fretboard.SetTuning(m.tuning)
	}

	// Update tablature display with new tuning
	m.updateTablatureConfig()
}

// trackHasLyrics checks if the track has any lyrics (in sections or at track level)
func (m *TUIModel) trackHasLyrics() bool {
	if m.track == nil {
		return false
	}
	// Check section-level beat-mapped lyrics
	for _, section := range m.track.Sections {
		if section.Lyrics != "" {
			return true
		}
	}
	// Check track-level per-bar lyrics
	for _, lyric := range m.track.Lyrics {
		if lyric != "" {
			return true
		}
	}
	return false
}

// trackCanEdit checks if the track can be edited
// Always allow edit mode - user can create new sections with Ctrl+N
func (m *TUIModel) trackCanEdit() bool {
	return m.track != nil
}

// isSixteenthNoteStyle checks if current style uses 16th notes
func (m *TUIModel) isSixteenthNoteStyle() bool {
	if m.track.Rhythm == nil {
		return false
	}
	switch m.track.Rhythm.Style {
	case "sixteenth", "funk_16th", "funk_muted", "dust_in_wind", "landslide", "pima", "pima_reverse":
		return true
	}
	return false
}

// isFingerPickingStyle checks if current style is fingerpicking
func (m *TUIModel) isFingerPickingStyle() bool {
	if m.track.Rhythm == nil {
		return false
	}
	style := m.track.Rhythm.Style
	return style == "fingerpick" || style == "fingerpick_slow" ||
		style == "travis" || style == "arpeggio_up" || style == "arpeggio_down"
}

// getUniqueChords returns unique chord symbols from the song
func (m *TUIModel) getUniqueChords() []string {
	seen := make(map[string]bool)
	var unique []string
	for _, bar := range m.bars {
		for _, bc := range bar.Chords {
			symbol := bc.Symbol
			if idx := strings.Index(symbol, "/"); idx > 0 {
				symbol = symbol[:idx]
			}
			if !seen[symbol] {
				seen[symbol] = true
				unique = append(unique, symbol)
			}
		}
	}
	return unique
}

// getStrumPatternSymbols returns the strum pattern as symbols
func (m *TUIModel) getStrumPatternSymbols() []string {
	if m.track.Rhythm == nil {
		return []string{"↓", ".", "↓", ".", "↓", ".", "↓", "."}
	}

	switch m.track.Rhythm.Style {
	case "fingerpick_slow":
		return []string{"↓", ".", ".", ".", "↓", ".", ".", "."}
	case "fingerpick", "travis":
		return []string{"↓", ".", "↑", ".", "↓", ".", "↑", "."}
	case "arpeggio_up", "arpeggio_down":
		return []string{"↓", "↓", "↓", "↓", "↓", "↓", "↓", "↓"}
	case "sixteenth":
		return []string{"↓", ".", "↑", ".", "↓", ".", "↑", ".", "↓", ".", "↑", ".", "↓", ".", "↑", "."}
	case "funk_16th":
		return []string{"↓", ".", "x", ".", "↑", "x", "↓", ".", "x", ".", "↑", ".", "↓", "x", "↑", "."}
	case "funk_muted":
		return []string{"x", ".", "↓", ".", "x", ".", "↑", ".", "x", ".", "↓", ".", "x", ".", "↑", "."}
	case "ska", "skank":
		return []string{".", "↓", ".", "↓", ".", "↓", ".", "↓"}
	case "reggae", "one_drop":
		return []string{".", ".", ".", ".", "↓", ".", ".", "."}
	case "country", "train":
		return []string{"↓", ".", "↓", ".", "↓", ".", "↓", "."}
	case "disco":
		return []string{"↓", ".", "↓", ".", "↓", ".", "↓", "."}
	case "motown", "soul":
		return []string{"↓", ".", "↓", "↑", "↓", ".", "↓", "↑"}
	case "flamenco", "rumba":
		return []string{"↓", ".", ".", "↓", ".", ".", "↓", ".", "↓", ".", "↓", ".", "↓", ".", ".", "."}
	default:
		return []string{"↓", ".", "↑", ".", "↓", ".", "↑", "."}
	}
}

// getPickingPattern returns the picking pattern tablature
func (m *TUIModel) getPickingPattern() []string {
	if m.track.Rhythm == nil {
		return []string{}
	}

	switch m.track.Rhythm.Style {
	case "fingerpick_slow":
		return []string{
			"e|----0-------0---|",
			"B|------0-------0-|",
			"G|--0-------0-----|",
			"D|----------------|",
			"A|----------------|",
			"E|0-------0-------|",
		}
	case "fingerpick":
		return []string{
			"e|----0---0---0---|",
			"B|------0---0---0-|",
			"G|--0---0---0---0-|",
			"D|----------------|",
			"A|----------------|",
			"E|0---0---0---0---|",
		}
	case "travis":
		return []string{
			"e|------0---0-----|",
			"B|----0---0---0---|",
			"G|--0-------0-----|",
			"D|----------------|",
			"A|----0-------0---|",
			"E|0-------0-------|",
		}
	case "arpeggio_up":
		// p-i-m-a: Bass, G, B, e, Bass, G, B, e (ascending treble)
		return []string{
			"e|------0-------0-|",
			"B|----0-------0---|",
			"G|--0-------0-----|",
			"D|----------------|",
			"A|----------------|",
			"E|0-------0-------|",
		}
	case "arpeggio_down":
		// p-a-m-i: Bass, e, B, G, Bass, e, B, G (descending treble)
		return []string{
			"e|--0-------0-----|",
			"B|----0-------0---|",
			"G|------0-------0-|",
			"D|----------------|",
			"A|----------------|",
			"E|0-------0-------|",
		}
	default:
		return []string{}
	}
}

// Stop signals the model to stop
func (m *TUIModel) Stop() {
	m.playing = false
}

// IsQuitting returns whether the user quit
func (m *TUIModel) IsQuitting() bool {
	return m.quitting
}

// getBeatsPerBar returns the number of beats per bar from the time signature
func (m *TUIModel) getBeatsPerBar() int {
	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar)
	}
	return beatsPerBar
}
