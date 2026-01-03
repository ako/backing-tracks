package display

import (
	"fmt"
	"strings"
	"time"

	"backing-tracks/midi"
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
			Align(lipgloss.Left)

	currentChordStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				Width(20).
				Align(lipgloss.Left)

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
	SetFingerstylePattern(pattern midi.PatternType)
	GetFingerstylePattern() midi.PatternType
	ToggleLoop(length int)                                 // Toggle loop of N bars from current position
	GetLoop() (enabled bool, startBar, endBar, length int) // Get loop state
	AdjustTempo(deltaBPM int)                              // Adjust playback tempo by delta BPM
	GetTempo() (effectiveBPM int, offset int)              // Get current effective tempo and offset
	GetCurrentSection() (name string, startBar, endBar int) // Get current section info
	LoopCurrentSection()                                    // Toggle loop for current section
	GetCurrentLyrics() (text string, chords []string)       // Get lyrics at current position
	GetLyricsForBar(bar int) (text string, chords []string) // Get lyrics for specific bar
	GetBeatLyricsForBar(bar int) []struct {                 // Get lyrics with beat positions
		Beat   int
		Lyrics string
		Chord  string
	}
	HasLyrics() bool                                        // Check if track has any lyrics
	UpdateLyrics(lyrics []parser.BeatLyric)                 // Update lyrics after editing
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
	quitting        bool

	// Display toggles (for left column components)
	showLyrics       bool // Show lyrics display
	showMetronome    bool // Show beat counter/metronome
	showStrumPattern bool // Show strum pattern visualization
	showTablature    bool // Show inline tablature
	showChordNames   bool // Show chord names above lyrics

	// Edit mode state (for lyrics timing adjustment)
	editMode      bool               // Whether in edit mode
	editBar       int                // Selected bar number
	editBeat      int                // Selected beat within bar (0-3)
	editLyrics    []parser.BeatLyric // Working copy of lyrics for editing
	editDirty     bool               // Whether changes have been made
	editFilename  string             // BTML filename for saving

	// Edit focus: "lyrics", "form", "sections"
	editFocus        string   // Current edit focus area
	editFormIndex    int      // Selected form entry index
	editSectionIndex int      // Selected section index
	editForm         []string // Working copy of form for editing
	editChordInput   string   // Current chord input buffer

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
		track:            track,
		bars:             bars,
		chords:           track.Progression.GetChords(),
		tempo:            track.Info.Tempo,
		timePerBeat:      timePerBeat,
		fretboard:        fretboard,
		chordChart:       chordChart,
		tablature:        tablature,
		currentScale:     scale,
		tuning:           tuning,
		tuningIndex:      tuningIndex,
		tuningName:       tuningName,
		capoPosition:     track.Info.Capo, // Initialize from track
		playing:          true,
		width:            120,
		height:           30,
		// Display toggles - sensible defaults
		showLyrics:       hasLyrics, // Enable by default if track has lyrics
		showMetronome:    true,      // Show beat counter by default
		showStrumPattern: true,      // Show strum pattern by default
		showTablature:    false,     // Inline tablature disabled by default
		showChordNames:   true,      // Show chord names by default
	}
}

// SetPlayer sets the audio player controller for synced playback
func (m *TUIModel) SetPlayer(p PlayerController) {
	m.player = p
}

// SetFilename sets the BTML filename for saving edits
func (m *TUIModel) SetFilename(filename string) {
	m.editFilename = filename
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
		// Handle edit mode keys first
		if m.editMode {
			beatsPerBar := 4
			key := msg.String()

			// Global edit mode keys (work regardless of focus)
			switch key {
			case "esc":
				// Discard changes and exit edit mode
				m.discardEdits()
				return m, nil
			case "ctrl+s":
				// Save changes and exit edit mode
				m.saveEdits()
				return m, nil
			case "ctrl+f":
				// Cycle focus: lyrics -> form -> sections -> lyrics
				m.cycleEditFocus()
				return m, nil
			}

			// Handle keys based on current focus
			switch m.editFocus {
			case "lyrics":
				return m.handleLyricsEditKeys(key, beatsPerBar)
			case "form":
				return m.handleFormEditKeys(key)
			case "sections":
				return m.handleSectionsEditKeys(key)
			}

			// Ignore other keys in edit mode
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "e":
			// Toggle edit mode (if track has editable content)
			if m.trackCanEdit() {
				m.toggleEditMode()
			}
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
				m.showLyrics = !m.showLyrics
			}
		case "m":
			// Toggle metronome/beat counter
			m.showMetronome = !m.showMetronome
		case "s":
			// Toggle strum pattern
			m.showStrumPattern = !m.showStrumPattern
		case "c":
			// Toggle chord names display
			m.showChordNames = !m.showChordNames
		case "t":
			// Toggle inline tablature display in left column
			m.showTablature = !m.showTablature
		case ";":
			// Previous pattern type
			if m.tablature != nil {
				newPattern := m.tablature.PrevPattern()
				m.tablature.RegenerateTablature(m.track)
				// Also update the audio player
				if m.player != nil {
					m.player.SetFingerstylePattern(newPattern)
				}
			}
		case "'":
			// Next pattern type
			if m.tablature != nil {
				newPattern := m.tablature.NextPattern()
				m.tablature.RegenerateTablature(m.track)
				// Also update the audio player
				if m.player != nil {
					m.player.SetFingerstylePattern(newPattern)
				}
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

	// Three-column layout (left column is wider now)
	leftCol := m.renderLeftColumn()
	middleCol := m.renderMiddleColumn()
	rightCol := m.renderRightColumn()

	// Use wider style for left column
	wideColumnStyle := lipgloss.NewStyle().Padding(0, 1).Width(80)

	// Join columns horizontally
	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		wideColumnStyle.Render(leftCol),
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

	editIndicator := ""
	if m.editMode {
		editStyle := lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#FFFF00")).
			Foreground(lipgloss.Color("#000000"))
		focusStyle := lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#00AAFF")).
			Foreground(lipgloss.Color("#FFFFFF"))

		var focusText string
		switch m.editFocus {
		case "lyrics":
			focusText = focusStyle.Render(" LYRICS ")
		case "form":
			focusText = focusStyle.Render(" FORM ")
		case "sections":
			focusText = focusStyle.Render(" SECTIONS ")
		}
		editIndicator = editStyle.Render("  ✏ EDIT") + focusText
	} else {
		// Show hint if edit is available
		if m.trackCanEdit() {
			editIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [Press 'e' to edit]")
		}
	}

	return fmt.Sprintf("  %s    %s%s%s%s%s%s%s%s%s%s", title, info, sectionIndicator, capoIndicator, transposeIndicator, tuningIndicator, muteIndicator, scaleName, loopIndicator, pauseIndicator, editIndicator)
}

// renderLeftColumn renders the chord/beat display
func (m *TUIModel) renderLeftColumn() string {
	var lines []string

	// Adjust rows based on what's displayed (more components = fewer rows)
	maxRows := 10 // Default: chords, lyrics, beats - compact
	if m.showTablature {
		maxRows = 3 // Tablature takes 6+ lines per row
	} else if m.showStrumPattern {
		maxRows = 5 // Strum pattern takes more space
	}

	// In edit mode, follow the edit cursor; otherwise follow playback
	focusBar := m.currentBar
	if m.editMode {
		focusBar = m.editBar
	}

	startRow := focusBar / 2
	if startRow > 0 {
		startRow-- // Show previous row for context
	}

	for row := 0; row < maxRows; row++ {
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
	barWidth := 36 // Slightly wider for better readability
	gutterWidth := 8

	// Build gutter with bar numbers and section name
	barNum1 := startBar + 1
	barNum2 := startBar + 2
	gutterText := fmt.Sprintf("%d-%d", barNum1, barNum2)

	// Check if a section starts at either bar
	sectionName := ""
	if m.player != nil {
		for _, bar := range []int{startBar, startBar + 1} {
			name, sectionStart, _ := m.player.GetCurrentSection()
			// Check all sections for one starting at this bar
			sections := m.track.GetSectionInfos()
			for _, sec := range sections {
				if sec.StartBar == bar {
					sectionName = sec.Name
					break
				}
			}
			_ = name
			_ = sectionStart
		}
	} else {
		sections := m.track.GetSectionInfos()
		for _, sec := range sections {
			if sec.StartBar == startBar || sec.StartBar == startBar+1 {
				sectionName = sec.Name
				break
			}
		}
	}

	gutterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Width(gutterWidth)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Width(gutterWidth).
		Italic(true)

	// Build the info gutter (section name or bar numbers)
	var infoGutter string
	if sectionName != "" {
		if len(sectionName) > gutterWidth-1 {
			sectionName = sectionName[:gutterWidth-1]
		}
		infoGutter = sectionStyle.Render(sectionName)
	} else {
		infoGutter = gutterStyle.Render(gutterText)
	}

	// Empty gutter for subsequent lines
	emptyGutter := strings.Repeat(" ", gutterWidth)

	// Track if we've shown the info gutter yet
	shownInfoGutter := false

	// Line 1: Chord names (if enabled)
	if m.showChordNames {
		chordLine := infoGutter
		shownInfoGutter = true
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
	}

	// Line 2: Lyrics (if enabled and available)
	if m.showLyrics {
		// Use info gutter if not shown yet
		currentGutter := emptyGutter
		if !shownInfoGutter {
			currentGutter = infoGutter
			shownInfoGutter = true
		}

		if m.editMode {
			// Edit mode: show lyrics at beat positions with selected bar/beat highlighted
			lyricsLine := currentGutter
			beatsPerBar := 4

			for i := 0; i < 2; i++ {
				barIdx := startBar + i
				if barIdx >= len(m.bars) {
					continue
				}

				isBarSelected := (barIdx == m.editBar)

				// Build lyrics for this bar with beat-aligned words
				beatWidth := (barWidth - 4) / beatsPerBar
				barLyrics := ""

				for beat := 0; beat < beatsPerBar; beat++ {
					word := ""
					isBeatSelected := isBarSelected && (beat == m.editBeat)

					// Find the word at this bar/beat in editLyrics
					for j := range m.editLyrics {
						if m.editLyrics[j].Bar == barIdx && m.editLyrics[j].Beat == beat {
							word = m.editLyrics[j].Lyrics
							break
						}
					}

					// Show cursor position even if no word
					if word == "" && isBeatSelected {
						word = "_" // Cursor placeholder
					}

					// Pad word to beat width
					if len(word) > beatWidth-1 {
						word = word[:beatWidth-1]
					}
					for len(word) < beatWidth {
						word += " "
					}

					if isBeatSelected {
						// Highlight selected beat position
						selectedStyle := lipgloss.NewStyle().
							Bold(true).
							Background(lipgloss.Color("#FFFF00")).
							Foreground(lipgloss.Color("#000000"))
						barLyrics += selectedStyle.Render(strings.TrimRight(word, " "))
						padding := beatWidth - len(strings.TrimRight(word, " "))
						barLyrics += strings.Repeat(" ", padding)
					} else {
						barLyrics += word
					}
				}

				// Highlight entire bar if selected
				style := lyricsStyle.Width(barWidth)
				if isBarSelected {
					style = style.Bold(true).Background(lipgloss.Color("#333333"))
				}
				lyricsLine += style.Render(barLyrics)
			}

			// Always show lyrics line in edit mode (cursor needs to be visible)
			lines = append(lines, lyricsLine)
		} else {
			// Normal mode: show lyrics at beat positions
			lyricsLine := currentGutter
			hasAnyLyrics := false
			beatsPerBar := 4

			for i := 0; i < 2; i++ {
				barIdx := startBar + i
				if barIdx >= len(m.bars) {
					continue
				}

				// Build lyrics for this bar with beat-aligned words
				beatWidth := (barWidth - 4) / beatsPerBar
				barLyrics := ""

				// Get beat lyrics from player if available
				var beatLyricsForBar []struct {
					Beat   int
					Lyrics string
					Chord  string
				}
				if m.player != nil {
					beatLyricsForBar = m.player.GetBeatLyricsForBar(barIdx)
				}

				for beat := 0; beat < beatsPerBar; beat++ {
					word := ""

					// Find lyrics at this beat
					for _, bl := range beatLyricsForBar {
						if bl.Beat == beat && bl.Lyrics != "" {
							word = bl.Lyrics
							hasAnyLyrics = true
							break
						}
					}

					// Pad word to beat width
					if len(word) > beatWidth-1 {
						word = word[:beatWidth-1]
					}
					for len(word) < beatWidth {
						word += " "
					}
					barLyrics += word
				}

				style := lyricsStyle.Width(barWidth)
				if barIdx == m.currentBar {
					style = style.Bold(true)
				}
				lyricsLine += style.Render(barLyrics)
			}

			if hasAnyLyrics {
				lines = append(lines, lyricsLine)
			}
		}
	}

	// Line 3: Strum pattern (if enabled)
	if m.showStrumPattern {
		strumGutter := emptyGutter
		if !shownInfoGutter {
			strumGutter = infoGutter
			shownInfoGutter = true
		}
		strumLine := strumGutter
		for i := 0; i < 2; i++ {
			barIdx := startBar + i
			if barIdx < len(m.bars) {
				pattern := m.renderStrumPattern(barIdx == m.currentBar)
				strumLine += lipgloss.NewStyle().Width(barWidth).Render(pattern)
			}
		}
		lines = append(lines, strumLine)
	}

	// Line 4: Beat numbers/metronome (if enabled)
	if m.showMetronome {
		beatGutter := emptyGutter
		if !shownInfoGutter {
			beatGutter = infoGutter
			shownInfoGutter = true
		}
		beatLine := beatGutter
		for i := 0; i < 2; i++ {
			barIdx := startBar + i
			if barIdx < len(m.bars) {
				beats := m.renderBeatNumbers(barIdx == m.currentBar)
				beatLine += lipgloss.NewStyle().Width(barWidth).Render(beats)
			}
		}
		lines = append(lines, beatLine)
	}

	// Line 5+: Inline tablature (if enabled) - 6 strings
	if m.showTablature && m.tablature != nil {
		tabLines := m.renderInlineTablature(startBar, barWidth)
		if len(tabLines) > 0 {
			lines = append(lines, tabLines...)
		}
	}

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

// renderInlineTablature renders a 6-string tablature for 2 bars matching the width of other components
func (m *TUIModel) renderInlineTablature(startBar int, barWidth int) []string {
	if m.tablature == nil || m.tablature.tablature == nil {
		return nil
	}

	// Get tuning for string names
	stringNames := []string{"e", "B", "G", "D", "A", "E"}
	if len(m.tuning.Names) >= 6 {
		for i := 0; i < 6; i++ {
			stringNames[i] = m.tuning.Names[5-i] // Reverse: high to low
		}
	}

	tabStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	fretStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)

	// Build 6 lines, one per string
	var lines []string
	beatsPerBar := 8 // 8th note resolution

	// Calculate spacing to match bar width (barWidth includes some padding)
	// Each bar needs beatsPerBar positions plus separators
	posWidth := (barWidth - 4) / beatsPerBar // Width per position
	if posWidth < 2 {
		posWidth = 2
	}

	for stringIdx := 0; stringIdx < 6; stringIdx++ {
		actualString := 5 - stringIdx // Convert display index to internal (0=low E, 5=high e)

		// String name
		name := stringNames[stringIdx]
		if len(name) == 1 {
			name = name + " "
		}

		var barParts []string
		for barOffset := 0; barOffset < 2; barOffset++ {
			barIdx := startBar + barOffset
			isCurrentBar := barIdx == m.currentBar

			tabBar, _ := m.tablature.tablature.GetCurrentAndNextBars(barIdx)

			// Build the string line for this bar
			positions := make([]string, beatsPerBar)
			for i := range positions {
				positions[i] = ""
			}

			if tabBar != nil {
				for _, note := range tabBar.Notes {
					if note.String != actualString {
						continue
					}
					// Map beat (1-5) to position (0-7)
					pos := int((note.Beat - 1.0) * 2)
					if pos < 0 {
						pos = 0
					}
					if pos >= beatsPerBar {
						pos = beatsPerBar - 1
					}
					positions[pos] = fmt.Sprintf("%d", note.Fret)
				}
			}

			// Render positions with proper spacing
			var posParts []string
			for i, p := range positions {
				beatNum := i / 2 // Which quarter note (0-3)

				// Create padded position string
				var posStr string
				if p == "" {
					posStr = strings.Repeat("─", posWidth)
				} else {
					// Center the fret number in the position width
					padding := posWidth - len(p)
					leftPad := padding / 2
					rightPad := padding - leftPad
					posStr = strings.Repeat("─", leftPad) + p + strings.Repeat("─", rightPad)
				}

				if isCurrentBar && beatNum == m.currentBeat {
					posParts = append(posParts, activeStyle.Render(posStr))
				} else if p != "" {
					posParts = append(posParts, fretStyle.Render(posStr))
				} else {
					posParts = append(posParts, tabStyle.Render(posStr))
				}
			}
			barParts = append(barParts, strings.Join(posParts, ""))
		}

		line := fmt.Sprintf("%s├%s┼%s┤", tabStyle.Render(name), barParts[0], barParts[1])
		lines = append(lines, strings.Repeat(" ", 6)+line) // 6 = gutterWidth - 2 for string name
	}

	return lines
}

// renderMiddleColumn renders the scale fretboard and chord tones fretboard
// In edit mode, shows the song form and sections for editing
func (m *TUIModel) renderMiddleColumn() string {
	// In edit mode, show form and sections editor
	if m.editMode {
		return m.renderFormAndSectionsEditor()
	}

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

// renderFormAndSectionsEditor renders the form and sections editor in edit mode
func (m *TUIModel) renderFormAndSectionsEditor() string {
	var lines []string

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#FFFF00")).Foreground(lipgloss.Color("#000000"))
	focusedHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Form section header
	formHeaderText := " FORM"
	if m.editFocus == "form" {
		formHeaderText = focusedHeaderStyle.Render(" FORM") + " (+/- add/remove, Shift+↑/↓ reorder)"
	} else {
		formHeaderText = headerStyle.Render(" FORM")
	}
	lines = append(lines, formHeaderText)
	lines = append(lines, dimStyle.Render(" ─────────────────────────────────"))

	// Render form entries
	if len(m.editForm) == 0 {
		lines = append(lines, normalStyle.Render("   (no form defined)"))
	} else {
		for i, sectionName := range m.editForm {
			prefix := "   "
			if m.editFocus == "form" && i == m.editFormIndex {
				lines = append(lines, selectedStyle.Render(fmt.Sprintf(" ▶ %d. %s ", i+1, sectionName)))
			} else {
				lines = append(lines, normalStyle.Render(fmt.Sprintf("%s%d. %s", prefix, i+1, sectionName)))
			}
		}
	}

	lines = append(lines, "")

	// Sections header
	sectionsHeaderText := " SECTIONS"
	if m.editFocus == "sections" {
		sectionsHeaderText = focusedHeaderStyle.Render(" SECTIONS") + " (type chord, Space add, Del remove)"
	} else {
		sectionsHeaderText = headerStyle.Render(" SECTIONS")
	}
	lines = append(lines, sectionsHeaderText)
	lines = append(lines, dimStyle.Render(" ─────────────────────────────────"))

	// Render sections with chord progressions
	if len(m.track.Sections) == 0 {
		lines = append(lines, normalStyle.Render("   (no sections defined)"))
	} else {
		for i, section := range m.track.Sections {
			// Section name
			sectionPrefix := "   "
			if m.editFocus == "sections" && i == m.editSectionIndex {
				lines = append(lines, selectedStyle.Render(fmt.Sprintf(" ▶ %s ", section.Name)))
			} else {
				lines = append(lines, normalStyle.Render(fmt.Sprintf("%s%s", sectionPrefix, section.Name)))
			}

			// Show chord progression with wrapping
			chords := section.Progression.GetChords()
			if len(chords) > 0 {
				var chordNames []string
				for _, c := range chords {
					chordNames = append(chordNames, c.Symbol)
				}

				// Wrap chords to multiple lines (max ~30 chars per line)
				maxLineWidth := 30
				indent := "     "
				var chordLines []string
				currentLine := ""

				for _, chord := range chordNames {
					if currentLine == "" {
						currentLine = chord
					} else if len(currentLine)+1+len(chord) <= maxLineWidth {
						currentLine += " " + chord
					} else {
						chordLines = append(chordLines, currentLine)
						currentLine = chord
					}
				}
				if currentLine != "" {
					chordLines = append(chordLines, currentLine)
				}

				// Highlight if editing this section
				isSelected := m.editFocus == "sections" && i == m.editSectionIndex

				for j, chordLine := range chordLines {
					displayLine := indent + chordLine
					// Show chord input on the last line if editing
					if isSelected && j == len(chordLines)-1 && m.editChordInput != "" {
						displayLine = indent + chordLine + " + " + selectedStyle.Render(m.editChordInput+"_")
					}
					if isSelected {
						lines = append(lines, focusedHeaderStyle.Render(displayLine))
					} else {
						lines = append(lines, dimStyle.Render(displayLine))
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n")
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

// trackCanEdit checks if the track has content that can be edited (lyrics, sections, or form)
func (m *TUIModel) trackCanEdit() bool {
	if m.track == nil {
		return false
	}
	// Can edit if there are sections
	if len(m.track.Sections) > 0 {
		return true
	}
	// Can edit if there's a form
	if len(m.track.Form) > 0 {
		return true
	}
	// Can edit if there are lyrics
	return m.trackHasLyrics()
}

// toggleEditMode enters or exits lyrics edit mode
func (m *TUIModel) toggleEditMode() {
	if m.editMode {
		// Already in edit mode - exit without saving
		m.discardEdits()
		return
	}

	// Build the editLyrics from track's lyrics data
	m.editLyrics = m.collectAllLyrics()

	// Copy form for editing
	m.editForm = make([]string, len(m.track.Form))
	copy(m.editForm, m.track.Form)

	// Pause playback when entering edit mode
	if m.player != nil && !m.player.IsPaused() {
		m.player.TogglePause()
	}

	m.editMode = true
	m.editFocus = "lyrics" // Start with lyrics focus
	m.editBar = m.currentBar
	m.editBeat = 0
	m.editFormIndex = 0
	m.editSectionIndex = 0
	m.editChordInput = ""
	m.editDirty = false
}

// collectAllLyrics gathers all beat lyrics from all sections or track-level lyrics
func (m *TUIModel) collectAllLyrics() []parser.BeatLyric {
	var result []parser.BeatLyric

	// Get beats per bar from time signature
	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		if _, err := fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar); err != nil {
			beatsPerBar = 4
		}
	}

	// First try section-level beat-mapped lyrics
	sectionInfos := m.track.GetSectionInfos()
	for _, info := range sectionInfos {
		for i := range m.track.Sections {
			if m.track.Sections[i].Name == info.Name && m.track.Sections[i].Lyrics != "" {
				lyrics := parser.ParseBeatLyrics(m.track.Sections[i].Lyrics, info.StartBar, beatsPerBar)
				// Only include entries that have actual lyrics text
				for _, bl := range lyrics {
					if bl.Lyrics != "" {
						result = append(result, bl)
					}
				}
				break
			}
		}
	}

	// If no section lyrics, try track-level per-bar lyrics
	if len(result) == 0 && len(m.track.Lyrics) > 0 {
		for bar, lyric := range m.track.Lyrics {
			if lyric != "" {
				result = append(result, parser.BeatLyric{
					Bar:    bar,
					Beat:   0, // Per-bar lyrics start at beat 0
					Lyrics: lyric,
				})
			}
		}
	}

	return result
}

// moveWordEarlier moves the word at selected bar/beat one beat earlier
func (m *TUIModel) moveWordEarlier() {
	idx := m.getEditLyricIndex(m.editBar, m.editBeat)
	if idx < 0 {
		return
	}

	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar)
	}

	bl := &m.editLyrics[idx]

	// Move one beat earlier
	bl.Beat--
	if bl.Beat < 0 {
		bl.Beat = beatsPerBar - 1
		bl.Bar--
		if bl.Bar < 0 {
			// Can't go before bar 0
			bl.Bar = 0
			bl.Beat = 0
		}
	}

	m.editDirty = true
}

// moveWordLater moves the word at selected bar/beat one beat later
func (m *TUIModel) moveWordLater() {
	idx := m.getEditLyricIndex(m.editBar, m.editBeat)
	if idx < 0 {
		return
	}

	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar)
	}

	bl := &m.editLyrics[idx]

	// Move one beat later
	bl.Beat++
	if bl.Beat >= beatsPerBar {
		bl.Beat = 0
		bl.Bar++
	}

	m.editDirty = true
}

// moveWordsEarlierFromHere moves all words from current bar onwards one beat earlier
func (m *TUIModel) moveWordsEarlierFromHere() {
	if len(m.editLyrics) == 0 {
		return
	}

	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar)
	}

	// Move all words from current bar onwards
	for i := range m.editLyrics {
		bl := &m.editLyrics[i]
		if bl.Bar < m.editBar || (bl.Bar == m.editBar && bl.Beat < m.editBeat) {
			continue // Skip words before current position
		}
		bl.Beat--
		if bl.Beat < 0 {
			bl.Beat = beatsPerBar - 1
			bl.Bar--
			if bl.Bar < 0 {
				bl.Bar = 0
				bl.Beat = 0
			}
		}
	}

	m.editDirty = true
}

// moveWordsLaterFromHere moves all words from current bar onwards one beat later
func (m *TUIModel) moveWordsLaterFromHere() {
	if len(m.editLyrics) == 0 {
		return
	}

	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar)
	}

	// Move all words from current bar onwards
	for i := range m.editLyrics {
		bl := &m.editLyrics[i]
		if bl.Bar < m.editBar || (bl.Bar == m.editBar && bl.Beat < m.editBeat) {
			continue // Skip words before current position
		}
		bl.Beat++
		if bl.Beat >= beatsPerBar {
			bl.Beat = 0
			bl.Bar++
		}
	}

	m.editDirty = true
}

// saveEdits saves the edited lyrics back to the BTML file
func (m *TUIModel) saveEdits() {
	if !m.editDirty {
		m.editMode = false
		return
	}

	// Get beats per bar
	beatsPerBar := 4
	if m.track.Info.TimeSignature != "" {
		fmt.Sscanf(m.track.Info.TimeSignature, "%d/", &beatsPerBar)
	}

	// Check if we have section-level lyrics
	hasSectionLyrics := false
	for _, section := range m.track.Sections {
		if section.Lyrics != "" {
			hasSectionLyrics = true
			break
		}
	}

	if hasSectionLyrics || len(m.track.Sections) > 0 {
		// Save to section-level beat-mapped lyrics
		sectionInfos := m.track.GetSectionInfos()

		for i := range m.track.Sections {
			section := &m.track.Sections[i]

			// Find the section info
			var sectionStart, sectionEnd int
			for _, info := range sectionInfos {
				if info.Name == section.Name {
					sectionStart = info.StartBar
					sectionEnd = info.EndBar
					break
				}
			}

			// Collect edited lyrics for this section
			var sectionLyrics []parser.BeatLyric
			for _, bl := range m.editLyrics {
				if bl.Bar >= sectionStart && bl.Bar < sectionEnd {
					// Adjust bar to be relative to section start for serialization
					adjusted := bl
					adjusted.Bar -= sectionStart
					sectionLyrics = append(sectionLyrics, adjusted)
				}
			}

			// Serialize back to beat notation format (even for sections that didn't have lyrics before)
			if len(sectionLyrics) > 0 {
				section.Lyrics = parser.SerializeBeatLyrics(sectionLyrics, beatsPerBar)
			} else if section.Lyrics != "" {
				// Clear lyrics if all were removed
				section.Lyrics = ""
			}
		}
	} else {
		// Save to track-level per-bar lyrics
		// Find the max bar number to size the array
		maxBar := 0
		for _, bl := range m.editLyrics {
			if bl.Bar > maxBar {
				maxBar = bl.Bar
			}
		}

		// Create new lyrics array
		newLyrics := make([]string, maxBar+1)
		for _, bl := range m.editLyrics {
			// For per-bar format, we put the lyric at its bar position
			// If multiple words land on same bar, concatenate them
			if newLyrics[bl.Bar] != "" {
				newLyrics[bl.Bar] += " " + bl.Lyrics
			} else {
				newLyrics[bl.Bar] = bl.Lyrics
			}
		}
		m.track.Lyrics = newLyrics
	}

	// Save edited form back to track
	m.track.Form = m.editForm

	// Re-expand sections after form change (updates Progression)
	if len(m.track.Sections) > 0 && len(m.track.Form) > 0 {
		m.track.Progression.Pattern = "" // Clear to force re-expansion
		// Force re-expansion by calling LoadTrack logic
		m.expandTrackSections()
	}

	// Save the track to file
	if m.editFilename != "" {
		if err := parser.SaveTrack(m.track, m.editFilename); err != nil {
			// TODO: Show error to user
		}
	}

	// Update the player's lyrics so the display reflects the changes
	if m.player != nil {
		m.player.UpdateLyrics(m.editLyrics)
	}

	m.editMode = false
	m.editDirty = false
}

// discardEdits exits edit mode without saving
func (m *TUIModel) discardEdits() {
	m.editMode = false
	m.editLyrics = nil
	m.editDirty = false
}

// getEditLyricAt returns the edited lyrics at a specific bar/beat, or nil
func (m *TUIModel) getEditLyricAt(bar, beat int) *parser.BeatLyric {
	for i := range m.editLyrics {
		if m.editLyrics[i].Bar == bar && m.editLyrics[i].Beat == beat {
			return &m.editLyrics[i]
		}
	}
	return nil
}

// getEditLyricIndex returns the index of lyrics at a specific bar/beat, or -1 if not found
func (m *TUIModel) getEditLyricIndex(bar, beat int) int {
	for i := range m.editLyrics {
		if m.editLyrics[i].Bar == bar && m.editLyrics[i].Beat == beat {
			return i
		}
	}
	return -1
}

// isEditSelected returns true if the given bar/beat is currently selected
func (m *TUIModel) isEditSelected(bar, beat int) bool {
	if !m.editMode {
		return false
	}
	return m.editBar == bar && m.editBeat == beat
}

// isEditBarSelected returns true if the given bar is currently selected
func (m *TUIModel) isEditBarSelected(bar int) bool {
	if !m.editMode {
		return false
	}
	return m.editBar == bar
}

// cycleEditFocus cycles through the focus areas: lyrics -> form -> sections -> lyrics
func (m *TUIModel) cycleEditFocus() {
	switch m.editFocus {
	case "lyrics":
		m.editFocus = "form"
	case "form":
		m.editFocus = "sections"
	case "sections":
		m.editFocus = "lyrics"
	default:
		m.editFocus = "lyrics"
	}
}

// handleLyricsEditKeys handles keyboard input when focus is on lyrics
func (m *TUIModel) handleLyricsEditKeys(key string, beatsPerBar int) (tea.Model, tea.Cmd) {
	switch key {
	case "left":
		// Move to previous beat (wrap to previous bar)
		m.editBeat--
		if m.editBeat < 0 {
			if m.editBar > 0 {
				m.editBar--
				m.editBeat = beatsPerBar - 1
			} else {
				m.editBeat = 0
			}
		}
		return m, nil
	case "right":
		// Move to next beat (wrap to next bar)
		m.editBeat++
		if m.editBeat >= beatsPerBar {
			if m.editBar < len(m.bars)-1 {
				m.editBar++
				m.editBeat = 0
			} else {
				m.editBeat = beatsPerBar - 1
			}
		}
		return m, nil
	case "up":
		// Move to previous line (2 bars up)
		if m.editBar >= 2 {
			m.editBar -= 2
		} else {
			m.editBar = 0
		}
		return m, nil
	case "down":
		// Move to next line (2 bars down)
		if m.editBar < len(m.bars)-2 {
			m.editBar += 2
		} else {
			m.editBar = len(m.bars) - 1
		}
		return m, nil
	case "shift+left":
		// Move word at current position one beat earlier
		m.moveWordEarlier()
		return m, nil
	case "shift+right":
		// Move word at current position one beat later
		m.moveWordLater()
		return m, nil
	case "shift+tab":
		// Move all words from current bar one beat earlier
		m.moveWordsEarlierFromHere()
		return m, nil
	case "tab":
		// Move all words from current bar one beat later
		m.moveWordsLaterFromHere()
		return m, nil
	case "backspace":
		// Delete last character of word at selected bar/beat
		if idx := m.getEditLyricIndex(m.editBar, m.editBeat); idx >= 0 {
			word := m.editLyrics[idx].Lyrics
			if len(word) > 0 {
				m.editLyrics[idx].Lyrics = word[:len(word)-1]
				m.editDirty = true
			}
		}
		return m, nil
	case "delete":
		// Delete entire word at selected bar/beat
		if idx := m.getEditLyricIndex(m.editBar, m.editBeat); idx >= 0 {
			m.editLyrics[idx].Lyrics = ""
			m.editDirty = true
		}
		return m, nil
	case "enter":
		// Save and exit
		m.saveEdits()
		return m, nil
	default:
		// Handle printable characters - append to word at selected bar/beat
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			idx := m.getEditLyricIndex(m.editBar, m.editBeat)
			if idx >= 0 {
				// Append to existing word
				m.editLyrics[idx].Lyrics += key
				m.editDirty = true
			} else {
				// Create new word at this position
				m.editLyrics = append(m.editLyrics, parser.BeatLyric{
					Bar:    m.editBar,
					Beat:   m.editBeat,
					Lyrics: key,
				})
				m.editDirty = true
			}
			return m, nil
		}
	}
	return m, nil
}

// handleFormEditKeys handles keyboard input when focus is on form
func (m *TUIModel) handleFormEditKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		// Move selection up
		if m.editFormIndex > 0 {
			m.editFormIndex--
		}
		return m, nil
	case "down", "j":
		// Move selection down
		if m.editFormIndex < len(m.editForm)-1 {
			m.editFormIndex++
		}
		return m, nil
	case "K", "shift+up":
		// Move selected entry up in the form
		if m.editFormIndex > 0 && len(m.editForm) > 1 {
			m.editForm[m.editFormIndex], m.editForm[m.editFormIndex-1] = m.editForm[m.editFormIndex-1], m.editForm[m.editFormIndex]
			m.editFormIndex--
			m.editDirty = true
		}
		return m, nil
	case "J", "shift+down":
		// Move selected entry down in the form
		if m.editFormIndex < len(m.editForm)-1 && len(m.editForm) > 1 {
			m.editForm[m.editFormIndex], m.editForm[m.editFormIndex+1] = m.editForm[m.editFormIndex+1], m.editForm[m.editFormIndex]
			m.editFormIndex++
			m.editDirty = true
		}
		return m, nil
	case "+", "=", "a":
		// Add section to form - show picker for available sections
		m.showFormSectionPicker()
		return m, nil
	case "-", "d", "backspace", "delete":
		// Remove selected entry from form
		if len(m.editForm) > 0 && m.editFormIndex < len(m.editForm) {
			m.editForm = append(m.editForm[:m.editFormIndex], m.editForm[m.editFormIndex+1:]...)
			if m.editFormIndex >= len(m.editForm) && m.editFormIndex > 0 {
				m.editFormIndex--
			}
			m.editDirty = true
		}
		return m, nil
	case "enter":
		// Save and exit
		m.saveEdits()
		return m, nil
	default:
		// Number keys 1-9 to quickly add a section by index
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			sectionIdx := int(key[0] - '1')
			if sectionIdx < len(m.track.Sections) {
				sectionName := m.track.Sections[sectionIdx].Name
				// Insert after current position
				insertIdx := m.editFormIndex + 1
				if len(m.editForm) == 0 {
					insertIdx = 0
				}
				newForm := make([]string, 0, len(m.editForm)+1)
				newForm = append(newForm, m.editForm[:insertIdx]...)
				newForm = append(newForm, sectionName)
				newForm = append(newForm, m.editForm[insertIdx:]...)
				m.editForm = newForm
				m.editFormIndex = insertIdx
				m.editDirty = true
			}
			return m, nil
		}
	}
	return m, nil
}

// handleSectionsEditKeys handles keyboard input when focus is on sections
func (m *TUIModel) handleSectionsEditKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		// Move selection up
		if m.editSectionIndex > 0 {
			m.editSectionIndex--
			m.editChordInput = ""
		}
		return m, nil
	case "down":
		// Move selection down
		if m.editSectionIndex < len(m.track.Sections)-1 {
			m.editSectionIndex++
			m.editChordInput = ""
		}
		return m, nil
	case "backspace":
		// Remove last character from chord input
		if len(m.editChordInput) > 0 {
			m.editChordInput = m.editChordInput[:len(m.editChordInput)-1]
		}
		return m, nil
	case " ":
		// Space commits the chord and starts a new one
		if m.editChordInput != "" {
			m.addChordToSection(m.editChordInput)
			m.editChordInput = ""
		}
		return m, nil
	case "enter":
		// Commit any pending chord, then save and exit
		if m.editChordInput != "" {
			m.addChordToSection(m.editChordInput)
			m.editChordInput = ""
		}
		m.saveEdits()
		return m, nil
	case "delete":
		// Remove last chord from the selected section
		m.removeLastChordFromSection()
		return m, nil
	default:
		// Handle printable characters for chord input
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.editChordInput += key
			return m, nil
		}
	}
	return m, nil
}

// showFormSectionPicker cycles through available sections to add to the form
func (m *TUIModel) showFormSectionPicker() {
	if len(m.track.Sections) == 0 {
		return
	}
	// Add the first available section (or cycle through them)
	// For simplicity, just add the first section
	sectionName := m.track.Sections[0].Name
	insertIdx := m.editFormIndex + 1
	if len(m.editForm) == 0 {
		insertIdx = 0
	}
	newForm := make([]string, 0, len(m.editForm)+1)
	newForm = append(newForm, m.editForm[:insertIdx]...)
	newForm = append(newForm, sectionName)
	newForm = append(newForm, m.editForm[insertIdx:]...)
	m.editForm = newForm
	m.editFormIndex = insertIdx
	m.editDirty = true
}

// addChordToSection adds a chord to the current section's progression
func (m *TUIModel) addChordToSection(chord string) {
	if m.editSectionIndex < 0 || m.editSectionIndex >= len(m.track.Sections) {
		return
	}
	section := &m.track.Sections[m.editSectionIndex]
	currentPattern := string(section.Progression.Pattern)
	if currentPattern == "" {
		section.Progression.Pattern = parser.StringOrList(chord)
	} else {
		section.Progression.Pattern = parser.StringOrList(currentPattern + " " + chord)
	}
	m.editDirty = true
}

// removeLastChordFromSection removes the last chord from the current section's progression
func (m *TUIModel) removeLastChordFromSection() {
	if m.editSectionIndex < 0 || m.editSectionIndex >= len(m.track.Sections) {
		return
	}
	section := &m.track.Sections[m.editSectionIndex]
	currentPattern := string(section.Progression.Pattern)
	parts := strings.Fields(currentPattern)
	if len(parts) > 0 {
		parts = parts[:len(parts)-1]
		section.Progression.Pattern = parser.StringOrList(strings.Join(parts, " "))
		m.editDirty = true
	}
}

// expandTrackSections rebuilds the Progression from Sections and Form
func (m *TUIModel) expandTrackSections() {
	// Build a map of section name -> section
	sectionMap := make(map[string]*parser.Section)
	for i := range m.track.Sections {
		section := &m.track.Sections[i]
		sectionMap[section.Name] = section
		// Set defaults for each section
		if section.Progression.BarsPerChord == 0 {
			section.Progression.BarsPerChord = 1
		}
	}

	// Expand form into a single pattern
	var allChords []string
	for _, sectionName := range m.track.Form {
		section, ok := sectionMap[sectionName]
		if !ok {
			continue // Skip unknown sections
		}
		// Add section marker so GetChords() can track section boundaries
		allChords = append(allChords, "["+sectionName+"]")
		// Get chords from this section (without repeat applied)
		chords := section.Progression.GetChords()
		for _, chord := range chords {
			// Reconstruct the chord notation with duration
			if chord.Bars == 1.0 {
				allChords = append(allChords, chord.Symbol)
			} else {
				allChords = append(allChords, fmt.Sprintf("%s*%g", chord.Symbol, chord.Bars))
			}
		}
	}

	// Set the expanded progression
	m.track.Progression.Pattern = parser.StringOrList(strings.Join(allChords, " "))
	m.track.Progression.BarsPerChord = 1 // Already handled in the pattern
	m.track.Progression.Repeat = 1       // Form already specifies the structure

	// Rebuild bars from new progression
	m.bars = processChordsIntoBars(m.track)
	m.chords = m.track.Progression.GetChords()
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

	// Build toggle status display
	var toggles []string
	if m.showChordNames {
		toggles = append(toggles, "C")
	}
	if m.showLyrics {
		toggles = append(toggles, "L")
	}
	if m.showMetronome {
		toggles = append(toggles, "M")
	}
	if m.showStrumPattern {
		toggles = append(toggles, "S")
	}
	if m.showTablature {
		toggles = append(toggles, "T")
	}
	toggleStatus := ""
	if len(toggles) > 0 {
		toggleStatus = fmt.Sprintf(" [%s]", strings.Join(toggles, ""))
	}

	var controls1, controls2 string
	if m.editMode {
		// Edit mode help
		switch m.editFocus {
		case "lyrics":
			controls1 = headerStyle.Render("  [←/→] beat [↑/↓] line [type] edit text [Shift+←/→] move word [Tab/Shift+Tab] shift all from here")
		case "form":
			controls1 = headerStyle.Render("  [↑/↓] select [1-9/+] add section [-/Del] remove [Shift+↑/↓] reorder")
		case "sections":
			controls1 = headerStyle.Render("  [↑/↓] select section [type] chord name [Space] add chord [Del] remove last chord")
		}
		controls2 = headerStyle.Render("  [Ctrl+F] switch focus [Ctrl+S] save [Esc] cancel")
	} else {
		// Playback mode help
		controls1 = headerStyle.Render("  [space] pause [←/→] seek [↑/↓] transpose [Shift+↑/↓] tempo [[/]] capo [</>] tuning [1-5] mute")
		controls2 = headerStyle.Render("  [l]yrics [m]etro [s]trum [t]ab [c]hord [e]dit [;/'] pattern [Shift+1-9] loop [q]uit")
	}

	return fmt.Sprintf("  %s  %d%% (bar %d/%d)%s\n%s\n%s",
		progressStyle.Render(bar),
		int(progress*100),
		m.currentBar+1,
		len(m.bars),
		toggleStatus,
		controls1,
		controls2)
}

// Stop signals the model to stop
func (m *TUIModel) Stop() {
	m.playing = false
}

// IsQuitting returns whether the user quit
func (m *TUIModel) IsQuitting() bool {
	return m.quitting
}
