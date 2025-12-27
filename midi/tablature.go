package midi

import (
	"backing-tracks/parser"
	"backing-tracks/theory"
	"fmt"
	"strings"
)

// TablatureConfig holds configuration for tablature generation
type TablatureConfig struct {
	PatternType   PatternType
	Tuning        theory.Tuning
	Capo          int
	ShowFingers   bool // Show right-hand fingering (p, i, m, a)
	Complexity    string // "simple", "moderate", "advanced"
}

// DefaultTablatureConfig returns a sensible default configuration
func DefaultTablatureConfig() TablatureConfig {
	return TablatureConfig{
		PatternType: PatternArpeggio,
		Tuning:      theory.Tunings["standard"],
		Capo:        0,
		ShowFingers: true,
		Complexity:  "moderate",
	}
}

// Tablature represents the complete tablature for a song
type Tablature struct {
	Bars          []TabBar
	TimeSignature string
	Tempo         float64
	Config        TablatureConfig
}

// GenerateTablature creates tablature for an entire track
func GenerateTablature(track *parser.Track, config TablatureConfig) *Tablature {
	chords := track.Progression.GetChords()
	timeSignature := track.Info.TimeSignature
	if timeSignature == "" {
		timeSignature = "4/4"
	}

	// Get appropriate pattern based on style
	style := track.Info.Style
	if style == "" {
		style = "acoustic"
	}
	pattern := GetPatternForStyle(style, timeSignature)

	// Override with config pattern type if specified
	if config.PatternType != "" {
		pattern = GetPattern(config.PatternType, timeSignature)
	}

	var bars []TabBar
	barNum := 1

	for _, chord := range chords {
		// Generate bars for this chord (respecting bars_per_chord)
		numBars := int(chord.Bars)
		if numBars < 1 {
			numBars = 1
		}

		voicing := GetGuitarVoicing(chord.Symbol, config.Tuning)

		for i := 0; i < numBars; i++ {
			notes := ApplyPatternToVoicing(pattern, voicing, config.Tuning, config.Capo)
			bar := TabBar{
				ChordName: chord.Symbol,
				Notes:     notes,
				BarNumber: barNum,
			}
			bars = append(bars, bar)
			barNum++
		}
	}

	return &Tablature{
		Bars:          bars,
		TimeSignature: timeSignature,
		Tempo:         float64(track.Info.Tempo),
		Config:        config,
	}
}

// GetBarAt returns the bar at a specific position (0-indexed)
func (t *Tablature) GetBarAt(index int) *TabBar {
	if index < 0 || index >= len(t.Bars) {
		return nil
	}
	return &t.Bars[index]
}

// GetCurrentAndNextBars returns bars for display (current and lookahead)
func (t *Tablature) GetCurrentAndNextBars(currentBar int) (*TabBar, *TabBar) {
	var current, next *TabBar

	if currentBar >= 0 && currentBar < len(t.Bars) {
		current = &t.Bars[currentBar]
	}
	if currentBar+1 >= 0 && currentBar+1 < len(t.Bars) {
		next = &t.Bars[currentBar+1]
	}

	return current, next
}

// RenderBar renders a single bar as ASCII tablature
func (t *Tablature) RenderBar(bar *TabBar, width int) []string {
	if bar == nil {
		return []string{}
	}

	// Calculate positions for each beat
	beatsPerBar := 4 // Default 4/4
	if t.TimeSignature == "3/4" {
		beatsPerBar = 3
	} else if t.TimeSignature == "6/8" {
		beatsPerBar = 6
	}

	// Characters per beat (width / beats)
	charsPerBeat := (width - 4) / beatsPerBar
	if charsPerBeat < 2 {
		charsPerBeat = 2
	}

	// Initialize string lines with dashes
	stringLines := make([][]rune, 6)
	stringNames := []string{"e", "B", "G", "D", "A", "E"}

	for i := 0; i < 6; i++ {
		stringLines[i] = make([]rune, width)
		for j := 0; j < width; j++ {
			stringLines[i][j] = '─'
		}
	}

	// Place notes
	for _, note := range bar.Notes {
		stringIdx := 5 - note.String // Reverse for display (high e at top)
		if stringIdx < 0 || stringIdx >= 6 {
			continue
		}

		// Calculate position based on beat
		pos := int((note.Beat - 1.0) * float64(charsPerBeat))
		if pos < 0 {
			pos = 0
		}
		if pos >= width-2 {
			pos = width - 3
		}

		// Write fret number
		fretStr := fmt.Sprintf("%d", note.Fret)
		for j, c := range fretStr {
			if pos+j < width {
				stringLines[stringIdx][pos+j] = c
			}
		}
	}

	// Build output lines
	var lines []string
	for i := 0; i < 6; i++ {
		line := fmt.Sprintf("%s ├%s┤", stringNames[i], string(stringLines[i]))
		lines = append(lines, line)
	}

	return lines
}

// RenderTwoBarView renders current and next bar side by side
func (t *Tablature) RenderTwoBarView(currentBarIdx int, totalWidth int) []string {
	current, next := t.GetCurrentAndNextBars(currentBarIdx)

	barWidth := (totalWidth - 5) / 2 // Leave room for divider
	if barWidth < 10 {
		barWidth = 10
	}

	var lines []string
	stringNames := []string{"e", "B", "G", "D", "A", "E"}

	// Header with chord names
	currentName := ""
	nextName := ""
	if current != nil {
		currentName = current.ChordName
	}
	if next != nil {
		nextName = next.ChordName
	}

	header := fmt.Sprintf("  %-*s │ %-*s", barWidth, currentName, barWidth, nextName)
	lines = append(lines, header)

	// Get beat markers
	beatsPerBar := 4
	if t.TimeSignature == "3/4" {
		beatsPerBar = 3
	} else if t.TimeSignature == "6/8" {
		beatsPerBar = 6
	}

	// Render each string
	for stringIdx := 0; stringIdx < 6; stringIdx++ {
		currentLine := t.renderStringLine(current, stringIdx, barWidth, beatsPerBar)
		nextLine := t.renderStringLine(next, stringIdx, barWidth, beatsPerBar)

		line := fmt.Sprintf("%s ├%s┼%s┤", stringNames[stringIdx], currentLine, nextLine)
		lines = append(lines, line)
	}

	// Beat markers
	beatMarkers := t.renderBeatMarkers(barWidth, beatsPerBar)
	lines = append(lines, fmt.Sprintf("    %s   %s", beatMarkers, beatMarkers))

	return lines
}

// renderStringLine renders a single string line for one bar
func (t *Tablature) renderStringLine(bar *TabBar, displayStringIdx int, width int, beatsPerBar int) string {
	line := make([]rune, width)
	for i := 0; i < width; i++ {
		line[i] = '─'
	}

	if bar == nil {
		return string(line)
	}

	// displayStringIdx: 0=high e, 5=low E
	// note.String: 0=low E, 5=high e
	actualString := 5 - displayStringIdx

	charsPerBeat := width / beatsPerBar
	if charsPerBeat < 2 {
		charsPerBeat = 2
	}

	for _, note := range bar.Notes {
		if note.String != actualString {
			continue
		}

		// Calculate position
		pos := int((note.Beat - 1.0) * float64(charsPerBeat))
		if pos < 0 {
			pos = 0
		}
		if pos >= width-1 {
			pos = width - 2
		}

		// Write fret number
		fretStr := fmt.Sprintf("%d", note.Fret)
		for j, c := range fretStr {
			if pos+j < width {
				line[pos+j] = c
			}
		}
	}

	return string(line)
}

// renderBeatMarkers creates beat number markers
func (t *Tablature) renderBeatMarkers(width int, beatsPerBar int) string {
	charsPerBeat := width / beatsPerBar
	if charsPerBeat < 2 {
		charsPerBeat = 2
	}

	var markers strings.Builder
	for beat := 1; beat <= beatsPerBar; beat++ {
		marker := fmt.Sprintf("%d", beat)
		markers.WriteString(marker)
		// Pad to fill charsPerBeat
		for i := len(marker); i < charsPerBeat; i++ {
			markers.WriteRune(' ')
		}
	}

	result := markers.String()
	if len(result) > width {
		result = result[:width]
	}
	return result
}

// RenderPlayhead returns a playhead indicator for the current beat position
func (t *Tablature) RenderPlayhead(currentBeat float64, barWidth int) string {
	beatsPerBar := 4.0
	if t.TimeSignature == "3/4" {
		beatsPerBar = 3.0
	} else if t.TimeSignature == "6/8" {
		beatsPerBar = 6.0
	}

	// Get beat within current bar (1-based)
	beatInBar := currentBeat
	for beatInBar > beatsPerBar {
		beatInBar -= beatsPerBar
	}
	if beatInBar < 1 {
		beatInBar = 1
	}

	// Calculate position
	charsPerBeat := barWidth / int(beatsPerBar)
	pos := int((beatInBar - 1.0) * float64(charsPerBeat))
	if pos < 0 {
		pos = 0
	}
	if pos >= barWidth {
		pos = barWidth - 1
	}

	// Create playhead line
	line := make([]rune, barWidth+4)
	for i := range line {
		line[i] = ' '
	}
	line[pos+4] = '▲'

	return string(line)
}

// GetTotalBars returns the total number of bars
func (t *Tablature) GetTotalBars() int {
	return len(t.Bars)
}

// PatternTypeFromString converts a string to PatternType
func PatternTypeFromString(s string) PatternType {
	switch strings.ToLower(s) {
	case "travis":
		return PatternTravis
	case "arpeggio":
		return PatternArpeggio
	case "folk":
		return PatternFolk
	case "classical":
		return PatternClassical
	case "bossa", "bossa_nova":
		return PatternBossaNova
	case "ballad":
		return PatternBallad
	case "waltz":
		return PatternWaltz
	default:
		return PatternFingerpick
	}
}

// AllPatternTypes returns all available pattern types
var AllPatternTypes = []PatternType{
	PatternTravis,
	PatternArpeggio,
	PatternFolk,
	PatternClassical,
	PatternBossaNova,
	PatternBallad,
	PatternWaltz,
	PatternFingerpick,
}

// NextPatternType returns the next pattern type in the cycle
func NextPatternType(current PatternType) PatternType {
	for i, pt := range AllPatternTypes {
		if pt == current {
			return AllPatternTypes[(i+1)%len(AllPatternTypes)]
		}
	}
	return AllPatternTypes[0]
}

// PrevPatternType returns the previous pattern type in the cycle
func PrevPatternType(current PatternType) PatternType {
	for i, pt := range AllPatternTypes {
		if pt == current {
			if i == 0 {
				return AllPatternTypes[len(AllPatternTypes)-1]
			}
			return AllPatternTypes[i-1]
		}
	}
	return AllPatternTypes[0]
}
