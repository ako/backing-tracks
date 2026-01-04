package display

import (
	"fmt"
	"strings"

	"backing-tracks/theory"

	"github.com/charmbracelet/lipgloss"
)

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
		case "track":
			focusText = focusStyle.Render(" TRACK ")
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
		sectionsHeaderText = focusedHeaderStyle.Render(" SECTIONS") + " (←→ select, type to edit, Del remove)"
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
			// Section name with number (for form editing - press number to add)
			sectionNum := i + 1 // 1-based for user display
			if m.editFocus == "sections" && i == m.editSectionIndex {
				if m.editSectionRename {
					// Show rename input
					lines = append(lines, selectedStyle.Render(fmt.Sprintf(" ▶ %d. %s_ ", sectionNum, m.editSectionName)))
				} else {
					lines = append(lines, selectedStyle.Render(fmt.Sprintf(" ▶ %d. %s ", sectionNum, section.Name)))
				}
			} else {
				lines = append(lines, normalStyle.Render(fmt.Sprintf("   %d. %s", sectionNum, section.Name)))
			}

			// Show chord progression with wrapping and individual chord selection
			chords := section.Progression.GetChords()
			isSelected := m.editFocus == "sections" && i == m.editSectionIndex

			if len(chords) > 0 || isSelected {
				// Wrap chords to multiple lines (max ~30 chars per line)
				maxLineWidth := 30
				indent := "     "

				// Build lines with individual chord highlighting
				currentLine := indent
				currentLineLen := 0

				for chordIdx, c := range chords {
					chordText := c.Symbol
					// Include duration modifier if not 1.0
					if c.Bars != 1.0 {
						if c.Bars == float64(int(c.Bars)) {
							chordText = fmt.Sprintf("%s*%d", c.Symbol, int(c.Bars))
						} else {
							chordText = fmt.Sprintf("%s*%g", c.Symbol, c.Bars)
						}
					}
					chordLen := len(chordText)

					// Check if we need to wrap to a new line
					if currentLineLen > 0 && currentLineLen+1+chordLen > maxLineWidth {
						lines = append(lines, currentLine)
						currentLine = indent
						currentLineLen = 0
					}

					// Add space separator if not at start of line
					if currentLineLen > 0 {
						currentLine += " "
						currentLineLen++
					}

					// Highlight selected chord
					if isSelected && chordIdx == m.editChordIndex {
						// If editing this chord, show input buffer
						if m.editChordInput != "" {
							chordText = m.editChordInput + "_"
						} else {
							chordText = "[" + chordText + "]"
						}
						currentLine += selectedStyle.Render(chordText)
					} else if isSelected {
						currentLine += focusedHeaderStyle.Render(chordText)
					} else {
						currentLine += dimStyle.Render(chordText)
					}
					currentLineLen += chordLen
				}

				// Show input for adding new chord at end
				if isSelected && m.editChordIndex == -1 && m.editChordInput != "" {
					// Start a new line for the input to avoid overflow
					if currentLine != indent {
						lines = append(lines, currentLine)
						currentLine = indent
					}
					inputDisplay := "+" + m.editChordInput + "_"
					currentLine += selectedStyle.Render(inputDisplay)
				}

				// Add the last line
				if currentLine != indent {
					lines = append(lines, currentLine)
				} else if isSelected {
					// Show placeholder when section has no chords
					lines = append(lines, indent+dimStyle.Render("(no chords - type to add)"))
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// renderRightColumn renders the chord charts and picking pattern
// In edit mode, shows track properties at the top
func (m *TUIModel) renderRightColumn() string {
	var lines []string

	// In edit mode, show track properties at the top
	if m.editMode {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
		focusedHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
		selectedStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#FFFF00")).Foreground(lipgloss.Color("#000000"))
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

		// Header with focus indicator
		trackHeaderText := " TRACK PROPERTIES"
		if m.editFocus == "track" {
			trackHeaderText = focusedHeaderStyle.Render(" TRACK PROPERTIES") + " (↑/↓ select, type to edit)"
		} else {
			trackHeaderText = headerStyle.Render(" TRACK PROPERTIES")
		}
		lines = append(lines, trackHeaderText)
		lines = append(lines, dimStyle.Render(" ─────────────────────────────────"))

		// Define field labels and values
		type trackField struct {
			label string
			value string
		}
		fields := []trackField{
			{"Title:", m.track.Info.Title},
			{"Key:", m.track.Info.Key},
			{"Tempo:", fmt.Sprintf("%d", m.track.Info.Tempo)},
			{"Time:", m.track.Info.TimeSignature},
			{"Style:", m.track.Info.Style},
			{"Capo:", fmt.Sprintf("%d", m.track.Info.Capo)},
			{"Tuning:", m.track.Info.Tuning},
		}
		// Add rhythm/bass/drums
		rhythmVal := ""
		if m.track.Rhythm != nil {
			rhythmVal = m.track.Rhythm.Style
		}
		fields = append(fields, trackField{"Rhythm:", rhythmVal})

		bassVal := ""
		if m.track.Bass != nil {
			bassVal = m.track.Bass.Style
		}
		fields = append(fields, trackField{"Bass:", bassVal})

		drumsVal := ""
		if m.track.Drums != nil {
			drumsVal = m.track.Drums.Style
		}
		fields = append(fields, trackField{"Drums:", drumsVal})

		// Render each field
		for i, f := range fields {
			isSelected := m.editFocus == "track" && m.editTrackField == i
			displayValue := f.value
			if displayValue == "" || displayValue == "0" {
				displayValue = "(empty)"
			}

			if isSelected {
				// Show input buffer if typing, otherwise show current value
				if m.editTrackInput != "" {
					displayValue = m.editTrackInput + "_"
				} else {
					displayValue = "[" + displayValue + "]"
				}
				lines = append(lines, fmt.Sprintf(" %s %s", labelStyle.Render(f.label), selectedStyle.Render(displayValue)))
			} else {
				lines = append(lines, fmt.Sprintf(" %s %s", labelStyle.Render(f.label), valueStyle.Render(displayValue)))
			}
		}
		lines = append(lines, "")
	}

	// Hide picking pattern and chord charts in edit mode
	if !m.editMode {
		// Picking pattern (if fingerpicking style)
		if m.isFingerPickingStyle() {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Render(" Picking Pattern:"))
			for _, patternLine := range m.getPickingPattern() {
				lines = append(lines, " "+patternLine)
			}
			lines = append(lines, "")
		}
	}

	// Chord charts for unique chords - 3 per row (hidden in edit mode)
	if m.editMode {
		return strings.Join(lines, "\n")
	}

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
			controls1 = headerStyle.Render("  [↑/↓] section [←/→] chord [r] rename [type] chord [Space] add [Del] remove")
		case "track":
			controls1 = headerStyle.Render("  [↑/↓] select [←/→] cycle options [type] new value [Tab] next [Enter] save")
		}
		controls2 = headerStyle.Render("  [Ctrl+F] focus [Ctrl+N] new section [Ctrl+S] save [Esc] cancel")
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
