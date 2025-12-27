package display

import (
	"fmt"
	"strings"

	"backing-tracks/midi"
	"backing-tracks/parser"
	"backing-tracks/theory"

	"github.com/charmbracelet/lipgloss"
)

// TablatureDisplay manages the tablature visualization
type TablatureDisplay struct {
	tablature     *midi.Tablature
	config        midi.TablatureConfig
	enabled       bool
	currentBar    int
	currentBeat   float64
	width         int
}

// Tablature styles
var (
	tabHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	tabStringStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	tabFretStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	tabCurrentFretStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	tabPlayheadStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00"))

	tabBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(0, 1)
)

// NewTablatureDisplay creates a new tablature display
func NewTablatureDisplay(track *parser.Track, tuning theory.Tuning, capo int) *TablatureDisplay {
	config := midi.TablatureConfig{
		PatternType: midi.PatternArpeggio,
		Tuning:      tuning,
		Capo:        capo,
		ShowFingers: true,
		Complexity:  "moderate",
	}

	// Set pattern based on track style
	if track.Info.Style != "" {
		config.PatternType = getPatternForTrackStyle(track.Info.Style)
	}

	tablature := midi.GenerateTablature(track, config)

	return &TablatureDisplay{
		tablature:  tablature,
		config:     config,
		enabled:    false, // Disabled by default
		currentBar: 0,
		width:      80,
	}
}

// getPatternForTrackStyle maps track style to pattern type
func getPatternForTrackStyle(style string) midi.PatternType {
	style = strings.ToLower(style)
	switch {
	case strings.Contains(style, "blues") || strings.Contains(style, "country"):
		return midi.PatternTravis
	case strings.Contains(style, "classical") || strings.Contains(style, "spanish"):
		return midi.PatternClassical
	case strings.Contains(style, "bossa") || strings.Contains(style, "latin"):
		return midi.PatternBossaNova
	case strings.Contains(style, "ballad"):
		return midi.PatternBallad
	case strings.Contains(style, "waltz"):
		return midi.PatternWaltz
	case strings.Contains(style, "folk"):
		return midi.PatternFolk
	default:
		return midi.PatternArpeggio
	}
}

// SetEnabled enables or disables tablature display
func (td *TablatureDisplay) SetEnabled(enabled bool) {
	td.enabled = enabled
}

// IsEnabled returns whether tablature display is enabled
func (td *TablatureDisplay) IsEnabled() bool {
	return td.enabled
}

// Toggle toggles tablature display on/off
func (td *TablatureDisplay) Toggle() {
	td.enabled = !td.enabled
}

// SetWidth sets the display width
func (td *TablatureDisplay) SetWidth(width int) {
	td.width = width
}

// SetPosition updates the current playback position
func (td *TablatureDisplay) SetPosition(bar int, beat float64) {
	td.currentBar = bar
	td.currentBeat = beat
}

// SetPatternType changes the fingerstyle pattern
func (td *TablatureDisplay) SetPatternType(pt midi.PatternType) {
	td.config.PatternType = pt
	// Regenerate tablature with new pattern (would need track reference)
}

// NextPattern cycles to the next pattern type
func (td *TablatureDisplay) NextPattern() midi.PatternType {
	td.config.PatternType = midi.NextPatternType(td.config.PatternType)
	return td.config.PatternType
}

// PrevPattern cycles to the previous pattern type
func (td *TablatureDisplay) PrevPattern() midi.PatternType {
	td.config.PatternType = midi.PrevPatternType(td.config.PatternType)
	return td.config.PatternType
}

// GetPatternType returns the current pattern type
func (td *TablatureDisplay) GetPatternType() midi.PatternType {
	return td.config.PatternType
}

// UpdateConfig updates tuning and capo
func (td *TablatureDisplay) UpdateConfig(tuning theory.Tuning, capo int) {
	td.config.Tuning = tuning
	td.config.Capo = capo
}

// RegenerateTablature regenerates the tablature with current config
func (td *TablatureDisplay) RegenerateTablature(track *parser.Track) {
	td.tablature = midi.GenerateTablature(track, td.config)
}

// Render renders the tablature display
func (td *TablatureDisplay) Render() string {
	if !td.enabled || td.tablature == nil {
		return ""
	}

	var b strings.Builder

	// Header
	patternName := string(td.config.PatternType)
	header := tabHeaderStyle.Render(fmt.Sprintf("  Tablature [%s]", patternName))
	b.WriteString(header)
	b.WriteString("\n")

	// Get current and next bars
	current, next := td.tablature.GetCurrentAndNextBars(td.currentBar)

	// Calculate bar width
	barWidth := (td.width - 10) / 2
	if barWidth < 20 {
		barWidth = 20
	}

	// Chord names
	currentName := ""
	nextName := ""
	if current != nil {
		currentName = current.ChordName
	}
	if next != nil {
		nextName = next.ChordName
	}

	chordLine := fmt.Sprintf("    %-*s   %-*s", barWidth, currentName, barWidth, nextName)
	b.WriteString(tabHeaderStyle.Render(chordLine))
	b.WriteString("\n")

	// Render each string (high e to low E)
	// Get string names from tuning config (stored low to high, display high to low)
	stringNames := []string{"e", "B", "G", "D", "A", "E"} // default
	if len(td.config.Tuning.Names) >= 6 {
		// Reverse the tuning names for display (low-to-high becomes high-to-low)
		stringNames = make([]string, 6)
		for i := 0; i < 6; i++ {
			stringNames[i] = td.config.Tuning.Names[5-i]
		}
	}

	for stringIdx := 0; stringIdx < 6; stringIdx++ {
		currentLine := td.renderStringLine(current, stringIdx, barWidth)
		nextLine := td.renderStringLine(next, stringIdx, barWidth)

		line := fmt.Sprintf("%s ├%s┼%s┤",
			tabStringStyle.Render(stringNames[stringIdx]),
			currentLine,
			nextLine)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Beat markers
	beatsPerBar := 4
	if td.tablature.TimeSignature == "3/4" {
		beatsPerBar = 3
	} else if td.tablature.TimeSignature == "6/8" {
		beatsPerBar = 6
	}

	beatMarkers := td.renderBeatMarkers(barWidth, beatsPerBar)
	b.WriteString(fmt.Sprintf("    %s   %s", beatMarkers, beatMarkers))
	b.WriteString("\n")

	// Playhead
	playhead := td.renderPlayhead(barWidth, beatsPerBar)
	b.WriteString(playhead)
	b.WriteString("\n")

	// Controls hint
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("  [t] toggle tab  [;/'] change pattern  [p] complexity")
	b.WriteString(hint)

	return b.String()
}

// renderStringLine renders a single string line for one bar
func (td *TablatureDisplay) renderStringLine(bar *midi.TabBar, displayStringIdx int, width int) string {
	line := make([]rune, width)
	for i := 0; i < width; i++ {
		line[i] = '─'
	}

	if bar == nil {
		return tabStringStyle.Render(string(line))
	}

	// displayStringIdx: 0=high e, 5=low E
	// note.String: 0=low E, 5=high e
	actualString := 5 - displayStringIdx

	beatsPerBar := 4
	if td.tablature.TimeSignature == "3/4" {
		beatsPerBar = 3
	} else if td.tablature.TimeSignature == "6/8" {
		beatsPerBar = 6
	}

	charsPerBeat := width / beatsPerBar
	if charsPerBeat < 3 {
		charsPerBeat = 3
	}

	// Track which positions have notes for highlighting
	notePositions := make(map[int]midi.TabNote)

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

		notePositions[pos] = note

		// Write fret number
		fretStr := fmt.Sprintf("%d", note.Fret)
		for j, c := range fretStr {
			if pos+j < width {
				line[pos+j] = c
			}
		}
	}

	// Convert to styled string
	result := string(line)

	// Check if any notes are currently being played
	if bar.BarNumber == td.currentBar+1 {
		// This is the current bar - could highlight current beat
		currentBeatPos := int((td.currentBeat - 1.0) * float64(charsPerBeat))
		for pos, _ := range notePositions {
			// Highlight notes near current beat
			if pos >= currentBeatPos-1 && pos <= currentBeatPos+1 {
				// Note is being played now - would need more complex styling
			}
		}
	}

	return tabFretStyle.Render(result)
}

// renderBeatMarkers creates beat number markers
func (td *TablatureDisplay) renderBeatMarkers(width int, beatsPerBar int) string {
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
	return tabStringStyle.Render(result)
}

// renderPlayhead returns a playhead indicator for the current beat position
func (td *TablatureDisplay) renderPlayhead(barWidth int, beatsPerBar int) string {
	// Get beat within current bar (1-based)
	beatInBar := td.currentBeat
	for beatInBar > float64(beatsPerBar) {
		beatInBar -= float64(beatsPerBar)
	}
	if beatInBar < 1 {
		beatInBar = 1
	}

	// Calculate position
	charsPerBeat := barWidth / beatsPerBar
	pos := int((beatInBar - 1.0) * float64(charsPerBeat))
	if pos < 0 {
		pos = 0
	}
	if pos >= barWidth {
		pos = barWidth - 1
	}

	// Create playhead line
	line := make([]rune, barWidth*2+10)
	for i := range line {
		line[i] = ' '
	}
	line[pos+4] = '▲'

	return tabPlayheadStyle.Render(string(line))
}

// RenderCompact renders a compact single-line status
func (td *TablatureDisplay) RenderCompact() string {
	if !td.enabled {
		return ""
	}

	status := fmt.Sprintf("TAB: %s", td.config.PatternType)
	return tabHeaderStyle.Render(status)
}
