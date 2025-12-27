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

// PlayerController interface for controlling audio playback
type PlayerController interface {
	TogglePause()
	SeekRelative(bars int)
	GetPlaybackState() (bar int, beat int, strum int, paused bool)
	IsPaused() bool
	Transpose(semitones int)
	GetTranspose() int
	SetCapo(fret int)
	GetCapo() int
	ToggleTrackMute(track int) // 0=drums, 1=bass, 2=chords, 3=melody, 4=fingerstyle
	IsTrackMuted(track int) bool
	ToggleLoop(length int)                                 // Toggle loop of N bars from current position
	GetLoop() (enabled bool, startBar, endBar, length int) // Get loop state
	AdjustTempo(deltaBPM int)                              // Adjust playback tempo by delta BPM
	GetTempo() (effectiveBPM int, offset int)              // Get current effective tempo and offset
	GetCurrentSection() (name string, startBar, endBar int) // Get current section info
	LoopCurrentSection()                                    // Toggle loop for current section
	GetCurrentLyrics() (text string, chords []string)       // Get lyrics at current position
	GetLyricsForBar(bar int) (text string, chords []string) // Get lyrics for specific bar
	HasLyrics() bool                                        // Check if track has any lyrics
}

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
	tablature    *TablatureDisplay
	currentScale *theory.Scale
	tuning       theory.Tuning
	tuningIndex  int    // Index into theory.TuningNames
	tuningName   string // Current tuning name for display

	// Layout
	width  int
	height int

	// State
	playing         bool
	paused          bool
	pausedAt        time.Time
	pausedTotal     time.Duration
	seekOffset      time.Duration // For seeking forward/backward
	transposeOffset int           // Semitones to transpose (+/-)
	capoPosition    int           // Capo fret position (0 = no capo)
	lyricsEnabled   bool          // Show lyrics display
	quitting        bool

	// Audio player (optional - for synced playback)
	player PlayerController
}

// NewTUIModel creates a new TUI model
func NewTUIModel(track *parser.Track) *TUIModel {
	beatsPerSecond := float64(track.Info.Tempo) / 60.0
	timePerBeat := time.Duration(float64(time.Second) / beatsPerSecond)

	bars := processChordsIntoBars(track)
	scale := theory.GetScaleForStyle(track.Info.Key, track.Info.Style, "")
	tuningName := track.Info.Tuning
	if tuningName == "" {
		tuningName = "standard"
	}
	tuning := theory.GetTuning(tuningName)
	tuningIndex := theory.GetTuningIndex(tuningName)
	fretboard := NewFretboardDisplayWithTuning(scale, 15, tuning)
	fretboard.SetCompactMode(true)
	chordChart := NewChordChart()
	tablature := NewTablatureDisplay(track, tuning, track.Info.Capo)

	// Check if track has lyrics (in sections or per-bar)
	hasLyrics := len(track.Lyrics) > 0
	for _, section := range track.Sections {
		if section.Lyrics != "" {
			hasLyrics = true
			break
		}
	}

	return &TUIModel{
		track:         track,
		bars:          bars,
		chords:        track.Progression.GetChords(),
		tempo:         track.Info.Tempo,
		timePerBeat:   timePerBeat,
		fretboard:     fretboard,
		chordChart:    chordChart,
		tablature:     tablature,
		currentScale:  scale,
		tuning:        tuning,
		tuningIndex:   tuningIndex,
		tuningName:    tuningName,
		capoPosition:  track.Info.Capo, // Initialize from track
		lyricsEnabled: hasLyrics,       // Enable by default if track has lyrics
		playing:       true,
		width:         120,
		height:        30,
	}
}

// SetPlayer sets the audio player controller for synced playback
func (m *TUIModel) SetPlayer(p PlayerController) {
	m.player = p
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
		case " ":
			// Toggle pause
			if m.player != nil {
				m.player.TogglePause()
			} else {
				if m.paused {
					m.pausedTotal += time.Since(m.pausedAt)
					m.paused = false
				} else {
					m.pausedAt = time.Now()
					m.paused = true
				}
			}
		case "left":
			// Jump to previous bar
			if m.player != nil {
				m.player.SeekRelative(-1)
			} else {
				timePerBar := m.timePerBeat * 4
				if m.currentBar > 0 {
					m.seekOffset -= timePerBar
				}
			}
		case "right":
			// Jump to next bar
			if m.player != nil {
				m.player.SeekRelative(1)
			} else {
				timePerBar := m.timePerBeat * 4
				if m.currentBar < len(m.bars)-1 {
					m.seekOffset += timePerBar
				}
			}
		case "up":
			// Transpose up one semitone
			if m.player != nil {
				m.player.Transpose(1)
				m.transposeOffset = m.player.GetTranspose()
			} else {
				m.transposeOffset++
			}
			m.updateTransposedScale()
		case "down":
			// Transpose down one semitone
			if m.player != nil {
				m.player.Transpose(-1)
				m.transposeOffset = m.player.GetTranspose()
			} else {
				m.transposeOffset--
			}
			m.updateTransposedScale()
		case "1":
			// Toggle drums
			if m.player != nil {
				m.player.ToggleTrackMute(0)
			}
		case "2":
			// Toggle bass
			if m.player != nil {
				m.player.ToggleTrackMute(1)
			}
		case "3":
			// Toggle chords
			if m.player != nil {
				m.player.ToggleTrackMute(2)
			}
		case "4":
			// Toggle melody
			if m.player != nil {
				m.player.ToggleTrackMute(3)
			}
		case "5":
			// Toggle fingerstyle
			if m.player != nil {
				m.player.ToggleTrackMute(4)
			}
		case "[":
			// Move capo down (with audio transpose)
			if m.capoPosition > 0 {
				m.capoPosition--
				if m.player != nil {
					m.player.SetCapo(m.capoPosition)
				}
				m.updateTablatureConfig()
			}
		case "]":
			// Move capo up (with audio transpose)
			if m.capoPosition < 12 {
				m.capoPosition++
				if m.player != nil {
					m.player.SetCapo(m.capoPosition)
				}
				m.updateTablatureConfig()
			}
		case "{":
			// Move capo down (visual only, no audio transpose)
			if m.capoPosition > 0 {
				m.capoPosition--
				m.updateTablatureConfig()
			}
		case "}":
			// Move capo up (visual only, no audio transpose)
			if m.capoPosition < 12 {
				m.capoPosition++
				m.updateTablatureConfig()
			}
		case ",", "<":
			// Previous tuning
			m.cycleTuning(-1)
		case ".", ">":
			// Next tuning
			m.cycleTuning(1)
		case "!":
			// Loop 1 bar (Shift+1)
			if m.player != nil {
				m.player.ToggleLoop(1)
			}
		case "@":
			// Loop 2 bars (Shift+2)
			if m.player != nil {
				m.player.ToggleLoop(2)
			}
		case "#":
			// Loop 3 bars (Shift+3)
			if m.player != nil {
				m.player.ToggleLoop(3)
			}
		case "$":
			// Loop 4 bars (Shift+4)
			if m.player != nil {
				m.player.ToggleLoop(4)
			}
		case "%":
			// Loop 5 bars (Shift+5)
			if m.player != nil {
				m.player.ToggleLoop(5)
			}
		case "^":
			// Loop 6 bars (Shift+6)
			if m.player != nil {
				m.player.ToggleLoop(6)
			}
		case "&":
			// Loop 7 bars (Shift+7)
			if m.player != nil {
				m.player.ToggleLoop(7)
			}
		case "*":
			// Loop 8 bars (Shift+8)
			if m.player != nil {
				m.player.ToggleLoop(8)
			}
		case "(":
			// Loop 9 bars (Shift+9)
			if m.player != nil {
				m.player.ToggleLoop(9)
			}
		case "shift+up":
			// Increase tempo by 5 BPM
			if m.player != nil {
				m.player.AdjustTempo(5)
			}
		case "shift+down":
			// Decrease tempo by 5 BPM
			if m.player != nil {
				m.player.AdjustTempo(-5)
			}
		case ")":
			// Loop current section (Shift+0)
			if m.player != nil {
				m.player.LoopCurrentSection()
			}
		case "l":
			// Toggle lyrics display
			if m.player != nil && m.player.HasLyrics() {
				m.lyricsEnabled = !m.lyricsEnabled
			}
		case "t":
			// Toggle tablature display
			if m.tablature != nil {
				m.tablature.Toggle()
			}
		case ";":
			// Previous pattern type
			if m.tablature != nil {
				m.tablature.PrevPattern()
			}
		case "'":
			// Next pattern type
			if m.tablature != nil {
				m.tablature.NextPattern()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		if m.playing {
			// Always update when we have a player (it controls pause state)
			// Otherwise check local pause state
			if m.player != nil || !m.paused {
				m.updatePosition()
			}
			return m, tickCmd()
		}
	}

	return m, nil
}

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

	// Tablature display (if enabled)
	if m.tablature != nil && m.tablature.IsEnabled() {
		m.tablature.SetWidth(m.width)
		b.WriteString(m.tablature.Render())
		b.WriteString("\n\n")
	}

	// Progress bar
	b.WriteString(m.renderProgressBar())

	return b.String()
}

// renderHeader renders the title and track info
func (m *TUIModel) renderHeader() string {
	title := titleStyle.Render(m.track.Info.Title)

	// Show transposed key if transpose is active
	displayKey := m.track.Info.Key
	if m.transposeOffset != 0 {
		displayKey = transposeChord(m.track.Info.Key, m.transposeOffset)
	}

	// Get effective tempo (may differ from original if speed adjusted)
	displayTempo := m.track.Info.Tempo
	tempoOffset := 0
	if m.player != nil {
		displayTempo, tempoOffset = m.player.GetTempo()
	}

	// Format BPM display - show offset if tempo was changed
	bpmDisplay := fmt.Sprintf("%d BPM", displayTempo)
	if tempoOffset != 0 {
		sign := "+"
		if tempoOffset < 0 {
			sign = ""
		}
		bpmDisplay = fmt.Sprintf("%d BPM (%s%d)", displayTempo, sign, tempoOffset)
	}

	info := headerStyle.Render(fmt.Sprintf("%s | %s | %s",
		displayKey, bpmDisplay, m.track.Info.Style))

	// Show capo indicator
	capoIndicator := ""
	if m.capoPosition > 0 {
		capoIndicator = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00CCCC")).
			Render(fmt.Sprintf("  [Capo %d]", m.capoPosition))
	}

	// Show transpose indicator
	transposeIndicator := ""
	if m.transposeOffset != 0 {
		sign := "+"
		if m.transposeOffset < 0 {
			sign = ""
		}
		transposeIndicator = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF00FF")).
			Render(fmt.Sprintf("  [%s%d]", sign, m.transposeOffset))
	}

	// Show track mute status
	muteIndicator := ""
	if m.player != nil {
		trackNames := []string{"Dr", "Ba", "Ch", "Me", "Fi"}
		var mutedTracks []string
		for i := 0; i < 5; i++ {
			if m.player.IsTrackMuted(i) {
				mutedTracks = append(mutedTracks, trackNames[i])
			}
		}
		if len(mutedTracks) > 0 {
			muteIndicator = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6666")).
				Render(fmt.Sprintf("  [MUTE: %s]", strings.Join(mutedTracks, ",")))
		}
	}

	scaleName := ""
	if m.currentScale != nil {
		scaleName = headerStyle.Render(" │ Scale: " + m.currentScale.Name)
	}

	// Show tuning indicator
	tuningIndicator := ""
	if m.tuningName != "" && m.tuningName != "standard" {
		tuningIndicator = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#66FF66")).
			Render(fmt.Sprintf("  [%s]", m.tuningName))
	}

	// Show current section
	sectionIndicator := ""
	if m.player != nil {
		if name, _, _ := m.player.GetCurrentSection(); name != "" {
			sectionIndicator = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFAA00")).
				Render(fmt.Sprintf("  § %s", name))
		}
	}

	pauseIndicator := ""
	if m.paused || (m.player != nil && m.player.IsPaused()) {
		pauseIndicator = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6600")).
			Render("  ⏸ PAUSED")
	}

	loopIndicator := ""
	if m.player != nil {
		if enabled, startBar, endBar, _ := m.player.GetLoop(); enabled {
			loopIndicator = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF00FF")).
				Render(fmt.Sprintf("  🔁 LOOP %d-%d", startBar+1, endBar))
		}
	}

	return fmt.Sprintf("  %s    %s%s%s%s%s%s%s%s%s", title, info, sectionIndicator, capoIndicator, transposeIndicator, tuningIndicator, muteIndicator, scaleName, loopIndicator, pauseIndicator)
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

	// Line 2: Lyrics (only if enabled and available)
	if m.lyricsEnabled {
		lyricsLine := "  "
		hasAnyLyrics := false
		for i := 0; i < 2; i++ {
			barIdx := startBar + i
			if barIdx < len(m.bars) {
				// Get lyrics from player if available, otherwise from bar
				lyrics := ""
				if m.player != nil {
					lyrics, _ = m.player.GetLyricsForBar(barIdx)
				}
				if lyrics == "" {
					lyrics = m.bars[barIdx].Lyrics
				}
				if lyrics != "" {
					hasAnyLyrics = true
				}
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
		if hasAnyLyrics {
			lines = append(lines, lyricsLine)
		}
	}

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

// renderMiddleColumn renders the scale fretboard and chord tones fretboard
func (m *TUIModel) renderMiddleColumn() string {
	if m.fretboard == nil || m.currentScale == nil {
		return ""
	}

	var lines []string

	// Scale name with capo indicator
	scaleName := m.currentScale.Name
	if m.capoPosition > 0 {
		scaleName = fmt.Sprintf("%s (capo %d)", scaleName, m.capoPosition)
	}
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(" "+scaleName))
	lines = append(lines, "")

	// Fret numbers (use 3-char columns for proper alignment with double digits)
	// Highlight the capo position
	fretLine := "   "
	for fret := 0; fret <= 12; fret++ {
		if fret == m.capoPosition && m.capoPosition > 0 {
			// Highlight capo position
			fretLine += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00CCCC")).Render(fmt.Sprintf("%2d ", fret))
		} else {
			fretLine += fmt.Sprintf("%2d ", fret)
		}
	}
	lines = append(lines, fretLine)

	// Strings (high to low) - use capo-adjusted tuning for positions
	tuning := m.getCapoAdjustedTuning()
	if len(tuning.Names) == 0 {
		tuning = theory.GetTuning("standard")
	}
	numStrings := len(tuning.Names)
	positions, roots := m.currentScale.GetFretboardPositionsWithTuning(12, tuning)

	for idx := 0; idx < numStrings; idx++ {
		stringIdx := numStrings - 1 - idx // Reverse order (high to low)
		name := tuning.Names[stringIdx]
		// Pad name for alignment
		if len(name) == 1 {
			name = " " + name
		}
		line := fmt.Sprintf("%s ", name)

		for fret := 0; fret <= 12; fret++ {
			if roots[stringIdx][fret] {
				line += lipgloss.NewStyle().Foreground(rootColor).Render(" ◆ ")
			} else if positions[stringIdx][fret] {
				line += lipgloss.NewStyle().Foreground(accentColor).Render(" ● ")
			} else {
				line += " · "
			}
		}
		lines = append(lines, line)
	}

	// Fret markers
	markerLine := "   "
	for fret := 0; fret <= 12; fret++ {
		if fret == 3 || fret == 5 || fret == 7 || fret == 9 {
			markerLine += " · "
		} else if fret == 12 {
			markerLine += " : "
		} else {
			markerLine += "   "
		}
	}
	lines = append(lines, markerLine)

	// Add chord tones fretboard
	chordLines := m.renderChordTonesFretboard()
	if len(chordLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, chordLines...)
	}

	return strings.Join(lines, "\n")
}

// renderChordTonesFretboard renders a fretboard showing all positions for current chord tones
func (m *TUIModel) renderChordTonesFretboard() []string {
	// Get current chord
	currentChord := m.getCurrentChordSymbol()
	if currentChord == "" {
		return nil
	}

	var lines []string

	// Chord name header
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(" "+currentChord+" Chord Tones"))
	lines = append(lines, "")

	// Get chord tones (returns slice of MIDI note offsets 0-11)
	chordTones := theory.GetChordTones(currentChord)
	if len(chordTones) == 0 {
		return nil
	}

	// Create a map for quick lookup
	toneMap := make(map[int]bool)
	for _, tone := range chordTones {
		toneMap[tone] = true
	}

	// Root note for highlighting
	rootTone := chordTones[0]

	// Use capo-adjusted tuning for positions
	tuning := m.getCapoAdjustedTuning()
	if len(tuning.Names) == 0 {
		tuning = theory.GetTuning("standard")
	}
	numStrings := len(tuning.Notes)

	// Fret numbers
	fretLine := "   "
	for fret := 0; fret <= 12; fret++ {
		fretLine += fmt.Sprintf("%2d ", fret)
	}
	lines = append(lines, fretLine)

	// Strings (high to low for display)
	for idx := 0; idx < numStrings; idx++ {
		stringIdx := numStrings - 1 - idx // Reverse to match display order
		openNote := tuning.Notes[stringIdx]
		name := tuning.Names[stringIdx]
		// Pad name for alignment
		if len(name) == 1 {
			name = " " + name
		}
		line := fmt.Sprintf("%s ", name)

		for fret := 0; fret <= 12; fret++ {
			noteAtFret := (openNote + fret) % 12
			if noteAtFret == rootTone {
				// Root note - highlight in different color
				line += lipgloss.NewStyle().Foreground(rootColor).Render(" ◆ ")
			} else if toneMap[noteAtFret] {
				// Chord tone
				line += lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(" ● ") // Orange for chord tones
			} else {
				line += " · "
			}
		}
		lines = append(lines, line)
	}

	// Fret markers
	markerLine := "   "
	for fret := 0; fret <= 12; fret++ {
		if fret == 3 || fret == 5 || fret == 7 || fret == 9 {
			markerLine += " · "
		} else if fret == 12 {
			markerLine += " : "
		} else {
			markerLine += "   "
		}
	}
	lines = append(lines, markerLine)

	return lines
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

	// Get current chord for highlighting (strip slash bass note for comparison)
	currentChord := m.getCurrentChordSymbol()
	if idx := strings.Index(currentChord, "/"); idx > 0 {
		currentChord = currentChord[:idx]
	}

	for _, chord := range uniqueChords {
		// First apply transpose to get the actual chord being played
		transposedChord := chord
		if m.transposeOffset != 0 {
			transposedChord = transposeChord(chord, m.transposeOffset)
		}

		// Check if this is the active chord
		isActive := (chord == currentChord)

		// If capo is set, transpose chord DOWN to get the shape to play
		// e.g., G chord with capo 2 = play F shape (F + capo 2 = G sound)
		displayChord := transposedChord
		shapeChord := transposedChord
		if m.capoPosition > 0 {
			shapeChord = transposeChord(transposedChord, -m.capoPosition)
			displayChord = fmt.Sprintf("%s→%s", transposedChord, shapeChord)
		}

		voicings := m.chordChart.GetVoicingsForTuning(shapeChord, m.tuningName)
		if len(voicings) == 0 {
			continue
		}
		// Override the name to show both original and shape
		voicing := voicings[0]
		voicing.Name = displayChord
		allDiagrams = append(allDiagrams, m.renderChordDiagram(voicing, isActive))
	}

	// Arrange 4 per row
	chartsPerRow := 4
	chartWidth := 20

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
				// Pad to fixed width (use lipgloss.Width to handle ANSI codes)
				visualWidth := lipgloss.Width(cell)
				if visualWidth < chartWidth {
					cell = cell + strings.Repeat(" ", chartWidth-visualWidth)
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
func (m *TUIModel) renderChordDiagram(v ChordVoicing, isActive bool) []string {
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

	// Highlight active chord with color
	nameStyle := lipgloss.NewStyle().Bold(true)
	if isActive {
		nameStyle = nameStyle.Foreground(lipgloss.Color("212")).Background(lipgloss.Color("236"))
	}
	lines = append(lines, nameStyle.Render(fmt.Sprintf(" %s [%s] ", v.Name, tabStr)))

	// Determine fret range
	startFret := 1
	if v.BaseFret > 0 {
		startFret = v.BaseFret
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

	controls := headerStyle.Render("  [space] pause  [←/→] seek  [↑/↓] transpose  [Shift+↑/↓] tempo  [[/]] capo  [{/}] visual capo  [</>] tuning  [l] lyrics  [t] tab  [q] quit")

	return fmt.Sprintf("  %s  %d%% (bar %d/%d)%s",
		progressStyle.Render(bar),
		int(progress*100),
		m.currentBar+1,
		len(m.bars),
		controls)
}

// Stop signals the model to stop
func (m *TUIModel) Stop() {
	m.playing = false
}

// IsQuitting returns whether the user quit
func (m *TUIModel) IsQuitting() bool {
	return m.quitting
}
