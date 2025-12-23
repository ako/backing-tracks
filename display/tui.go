package display

import (
	"fmt"
	"strings"
	"time"

	"backing-tracks/parser"
	"backing-tracks/theory"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the TUI
var (
	// Colors
	primaryColor   = lipgloss.Color("#00FFFF") // Cyan
	secondaryColor = lipgloss.Color("#FFFF00") // Yellow
	accentColor    = lipgloss.Color("#00FF00") // Green
	dimColor       = lipgloss.Color("#666666") // Gray
	rootColor      = lipgloss.Color("#FF6666") // Red for root notes

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	chordStyle = lipgloss.NewStyle().
			Width(20).
			Align(lipgloss.Center)

	currentChordStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				Width(20).
				Align(lipgloss.Center)

	lyricsStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Width(20)

	beatStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	currentBeatStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	columnStyle = lipgloss.NewStyle().
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#444444"))

	progressStyle = lipgloss.NewStyle().
			Foreground(accentColor)
)

// TickMsg is sent on each tick for time updates
type TickMsg time.Time

// TUIModel is the Bubbletea model for live display
type TUIModel struct {
	track        *parser.Track
	bars         []Bar
	chords       []parser.Chord
	tempo        int
	timePerBeat  time.Duration
	startTime    time.Time
	currentBar   int
	currentBeat  int
	currentStrum int

	// Display components
	fretboard    *FretboardDisplay
	chordChart   *ChordChart
	currentScale *theory.Scale

	// Layout
	width  int
	height int

	// State
	playing bool
	quitting bool
}

// NewTUIModel creates a new TUI model
func NewTUIModel(track *parser.Track) *TUIModel {
	beatsPerSecond := float64(track.Info.Tempo) / 60.0
	timePerBeat := time.Duration(float64(time.Second) / beatsPerSecond)

	bars := processChordsIntoBars(track)
	scale := theory.GetScaleForStyle(track.Info.Key, track.Info.Style, "")
	fretboard := NewFretboardDisplay(scale, 15)
	fretboard.SetCompactMode(true)
	chordChart := NewChordChart()

	return &TUIModel{
		track:        track,
		bars:         bars,
		chords:       track.Progression.GetChords(),
		tempo:        track.Info.Tempo,
		timePerBeat:  timePerBeat,
		fretboard:    fretboard,
		chordChart:   chordChart,
		currentScale: scale,
		playing:      true,
		width:        120,
		height:       30,
	}
}

// Init initializes the model
func (m *TUIModel) Init() tea.Cmd {
	m.startTime = time.Now()
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

// tickCmd returns a command that ticks every 50ms
func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles messages
func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		if m.playing {
			m.updatePosition()
			return m, tickCmd()
		}
	}

	return m, nil
}

// updatePosition calculates current bar/beat from elapsed time
func (m *TUIModel) updatePosition() {
	elapsed := time.Since(m.startTime)
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
}

// View renders the TUI
func (m *TUIModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Three-column layout
	leftCol := m.renderLeftColumn()
	middleCol := m.renderMiddleColumn()
	rightCol := m.renderRightColumn()

	// Join columns horizontally
	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		columnStyle.Render(leftCol),
		borderStyle.Render(middleCol),
		borderStyle.Render(rightCol),
	)
	b.WriteString(row)
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(m.renderProgressBar())

	return b.String()
}

// renderHeader renders the title and track info
func (m *TUIModel) renderHeader() string {
	title := titleStyle.Render(m.track.Info.Title)
	info := headerStyle.Render(fmt.Sprintf("%s | %d BPM | %s",
		m.track.Info.Key, m.track.Info.Tempo, m.track.Info.Style))

	scaleName := ""
	if m.currentScale != nil {
		scaleName = headerStyle.Render(" │ Scale: " + m.currentScale.Name)
	}

	return fmt.Sprintf("  %s    %s%s", title, info, scaleName)
}

// renderLeftColumn renders the chord/beat display
func (m *TUIModel) renderLeftColumn() string {
	var lines []string

	// Show 4 rows of 2 bars each
	startRow := m.currentBar / 2
	if startRow > 0 {
		startRow-- // Show previous row for context
	}

	for row := 0; row < 4; row++ {
		barIdx := (startRow + row) * 2
		if barIdx >= len(m.bars) {
			break
		}

		lines = append(lines, m.renderBarRow(barIdx))
		lines = append(lines, "") // Spacer
	}

	return strings.Join(lines, "\n")
}

// renderBarRow renders a row of 2 bars
func (m *TUIModel) renderBarRow(startBar int) string {
	var lines []string
	barWidth := 34

	// Line 1: Chord names
	chordLine := "  "
	for i := 0; i < 2; i++ {
		barIdx := startBar + i
		if barIdx < len(m.bars) {
			chord := m.getBarChordName(barIdx)
			if barIdx == m.currentBar {
				chordLine += currentChordStyle.Width(barWidth).Render(chord)
			} else {
				chordLine += chordStyle.Width(barWidth).Render(chord)
			}
		}
	}
	lines = append(lines, chordLine)

	// Line 2: Lyrics
	lyricsLine := "  "
	for i := 0; i < 2; i++ {
		barIdx := startBar + i
		if barIdx < len(m.bars) {
			lyrics := m.bars[barIdx].Lyrics
			if len(lyrics) > barWidth-2 {
				lyrics = lyrics[:barWidth-2]
			}
			style := lyricsStyle.Width(barWidth)
			if barIdx == m.currentBar && lyrics != "" {
				style = style.Bold(true)
			}
			lyricsLine += style.Render(lyrics)
		}
	}
	lines = append(lines, lyricsLine)

	// Line 3: Strum pattern
	strumLine := "  "
	for i := 0; i < 2; i++ {
		barIdx := startBar + i
		if barIdx < len(m.bars) {
			pattern := m.renderStrumPattern(barIdx == m.currentBar)
			strumLine += lipgloss.NewStyle().Width(barWidth).Render(pattern)
		}
	}
	lines = append(lines, strumLine)

	// Line 4: Beat numbers
	beatLine := "  "
	for i := 0; i < 2; i++ {
		barIdx := startBar + i
		if barIdx < len(m.bars) {
			beats := m.renderBeatNumbers(barIdx == m.currentBar)
			beatLine += lipgloss.NewStyle().Width(barWidth).Render(beats)
		}
	}
	lines = append(lines, beatLine)

	// Separator
	lines = append(lines, "  "+strings.Repeat("─", barWidth*2))

	return strings.Join(lines, "\n")
}

// getBarChordName returns the chord name for a bar
func (m *TUIModel) getBarChordName(barIdx int) string {
	if barIdx >= len(m.bars) || len(m.bars[barIdx].Chords) == 0 {
		return ""
	}
	return m.bars[barIdx].Chords[0].Symbol
}

// renderStrumPattern renders the strum pattern for a bar
func (m *TUIModel) renderStrumPattern(isCurrent bool) string {
	pattern := m.getStrumPatternSymbols()
	var result []string

	// Use narrower spacing for 16th notes
	spacing := "   "
	if len(pattern) > 8 {
		spacing = " "
	}

	for i, p := range pattern {
		if isCurrent {
			if i == m.currentStrum {
				result = append(result, currentBeatStyle.Render("█"))
			} else if i < m.currentStrum {
				result = append(result, beatStyle.Render(p))
			} else {
				result = append(result, beatStyle.Render("░"))
			}
		} else {
			result = append(result, beatStyle.Render(p))
		}
	}

	return " " + strings.Join(result, spacing)
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

// renderBeatNumbers renders the beat numbers
func (m *TUIModel) renderBeatNumbers(isCurrent bool) string {
	if m.isSixteenthNoteStyle() {
		return m.renderBeatNumbers16th(isCurrent)
	}

	beats := []string{"1", "2", "3", "4"}
	var result []string

	for i, b := range beats {
		if isCurrent && i == m.currentBeat {
			result = append(result, currentBeatStyle.Render("●"))
		} else if isCurrent && i == 0 {
			result = append(result, currentBeatStyle.Render("◉"))
		} else {
			result = append(result, beatStyle.Render(b))
		}
	}

	return " " + strings.Join(result, "       ")
}

// renderBeatNumbers16th renders beat numbers for 16th note patterns
func (m *TUIModel) renderBeatNumbers16th(isCurrent bool) string {
	// 16th note subdivisions: 1 e + a 2 e + a 3 e + a 4 e + a
	beats := []string{"1", "e", "+", "a", "2", "e", "+", "a", "3", "e", "+", "a", "4", "e", "+", "a"}
	var result []string

	for i, b := range beats {
		beatNum := i / 4 // Which quarter note beat (0-3)
		if isCurrent {
			if beatNum == m.currentBeat && i%4 == 0 {
				result = append(result, currentBeatStyle.Render("●"))
			} else if i == 0 && beatNum != m.currentBeat {
				result = append(result, currentBeatStyle.Render("◉"))
			} else {
				result = append(result, beatStyle.Render(b))
			}
		} else {
			result = append(result, beatStyle.Render(b))
		}
	}

	return " " + strings.Join(result, " ")
}

// renderMiddleColumn renders the scale fretboard
func (m *TUIModel) renderMiddleColumn() string {
	if m.fretboard == nil || m.currentScale == nil {
		return ""
	}

	var lines []string

	// Scale name
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(" "+m.currentScale.Name))
	lines = append(lines, "")

	// Fret numbers
	fretLine := "    "
	for fret := 0; fret <= 12; fret++ {
		if fret < 10 {
			fretLine += fmt.Sprintf("%d ", fret)
		} else {
			fretLine += fmt.Sprintf("%d", fret)
		}
	}
	lines = append(lines, fretLine)

	// Strings (high to low)
	stringNames := []string{"e", "B", "G", "D", "A", "E"}
	positions, roots := m.currentScale.GetFretboardPositions(12)

	for idx, name := range stringNames {
		stringIdx := 5 - idx // Reverse order
		line := fmt.Sprintf(" %s ", name)

		for fret := 0; fret <= 12; fret++ {
			if roots[stringIdx][fret] {
				line += lipgloss.NewStyle().Foreground(rootColor).Render("◆ ")
			} else if positions[stringIdx][fret] {
				line += lipgloss.NewStyle().Foreground(accentColor).Render("● ")
			} else {
				line += "· "
			}
		}
		lines = append(lines, line)
	}

	// Fret markers
	markerLine := "    "
	for fret := 0; fret <= 12; fret++ {
		if fret == 3 || fret == 5 || fret == 7 || fret == 9 {
			markerLine += "· "
		} else if fret == 12 {
			markerLine += ": "
		} else {
			markerLine += "  "
		}
	}
	lines = append(lines, markerLine)

	return strings.Join(lines, "\n")
}

// renderRightColumn renders the chord charts and picking pattern
func (m *TUIModel) renderRightColumn() string {
	var lines []string

	// Picking pattern (if fingerpicking style)
	if m.isFingerPickingStyle() {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(" Picking Pattern:"))
		for _, patternLine := range m.getPickingPattern() {
			lines = append(lines, " "+patternLine)
		}
		lines = append(lines, "")
	}

	// Chord charts for unique chords - 3 per row
	uniqueChords := m.getUniqueChords()
	var allDiagrams [][]string

	for _, chord := range uniqueChords {
		voicings := m.chordChart.GetVoicings(chord)
		if len(voicings) == 0 {
			continue
		}
		allDiagrams = append(allDiagrams, m.renderChordDiagram(voicings[0]))
	}

	// Arrange 3 per row
	chartsPerRow := 3
	chartWidth := 22

	for i := 0; i < len(allDiagrams); i += chartsPerRow {
		end := i + chartsPerRow
		if end > len(allDiagrams) {
			end = len(allDiagrams)
		}
		rowDiagrams := allDiagrams[i:end]

		// Find max height in this row
		maxHeight := 0
		for _, diag := range rowDiagrams {
			if len(diag) > maxHeight {
				maxHeight = len(diag)
			}
		}

		// Render row by joining diagrams horizontally
		for lineIdx := 0; lineIdx < maxHeight; lineIdx++ {
			var rowLine string
			for _, diag := range rowDiagrams {
				cell := ""
				if lineIdx < len(diag) {
					cell = diag[lineIdx]
				}
				// Pad to fixed width
				cellRunes := []rune(cell)
				if len(cellRunes) < chartWidth {
					cell = cell + strings.Repeat(" ", chartWidth-len(cellRunes))
				}
				rowLine += cell
			}
			lines = append(lines, rowLine)
		}
		lines = append(lines, "") // Spacer between rows
	}

	return strings.Join(lines, "\n")
}

// isSixteenthNoteStyle checks if current style uses 16th notes
func (m *TUIModel) isSixteenthNoteStyle() bool {
	if m.track.Rhythm == nil {
		return false
	}
	style := m.track.Rhythm.Style
	return style == "sixteenth" || style == "funk_16th" || style == "funk_muted"
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
	default:
		return []string{}
	}
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

// renderChordDiagram renders a single chord diagram
func (m *TUIModel) renderChordDiagram(v ChordVoicing) []string {
	var lines []string

	// Chord name and tab notation
	tabStr := ""
	for i := 0; i < 6; i++ {
		if v.Frets[i] == -1 {
			tabStr += "x"
		} else {
			tabStr += fmt.Sprintf("%d", v.Frets[i])
		}
	}
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf(" %s [%s]", v.Name, tabStr)))

	// String names
	lines = append(lines, " E  A  D  G  B  e")

	// Determine fret range
	startFret := 1
	if v.BaseFret > 0 {
		startFret = v.BaseFret
	}
	endFret := startFret + 3

	// Nut or fret indicator
	if startFret == 1 {
		lines = append(lines, " ══════════════════")
	} else {
		lines = append(lines, fmt.Sprintf(" %dfr─────────────", startFret))
	}

	// Frets
	for fret := startFret; fret <= endFret; fret++ {
		line := " "
		for str := 0; str < 6; str++ {
			f := v.Frets[str]
			if f == -1 && fret == startFret {
				line += "x  "
			} else if f == 0 && fret == startFret {
				line += "○  "
			} else if f == fret {
				line += "●  "
			} else {
				line += "│  "
			}
		}
		lines = append(lines, line)
	}

	return lines
}

// renderProgressBar renders the progress bar
func (m *TUIModel) renderProgressBar() string {
	progress := 0.0
	if len(m.bars) > 0 {
		progress = float64(m.currentBar) / float64(len(m.bars))
	}
	if progress > 1.0 {
		progress = 1.0
	}

	width := 50
	filled := int(progress * float64(width))
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)

	return fmt.Sprintf("  %s  %d%% (bar %d/%d)",
		progressStyle.Render(bar),
		int(progress*100),
		m.currentBar+1,
		len(m.bars))
}

// Stop signals the model to stop
func (m *TUIModel) Stop() {
	m.playing = false
}

// IsQuitting returns whether the user quit
func (m *TUIModel) IsQuitting() bool {
	return m.quitting
}
