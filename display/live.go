package display

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"backing-tracks/parser"
	"backing-tracks/theory"
)

// LiveDisplay shows real-time karaoke-style playback information
type LiveDisplay struct {
	track         *parser.Track
	chords        []parser.Chord
	bars          []Bar // Processed bars with chords and lyrics
	tempo         int
	timePerBeat   time.Duration
	startTime     time.Time
	stopChan      chan bool
	strumPattern  string
	barsPerLine   int
	displayLines  int
	fretboard     *FretboardDisplay
	currentScale  *theory.Scale
	showFretboard bool
	chordChart    *ChordChart
	currentChord  string
}

// Bar represents a single bar with its chords and lyrics
type Bar struct {
	Chords []BarChord // Chords in this bar (can be multiple for half-bar chords)
	Lyrics string     // Lyrics for this bar
}

// BarChord represents a chord within a bar
type BarChord struct {
	Symbol   string
	Beats    int     // Number of beats this chord occupies (1-4)
	StartBeat int    // Starting beat within the bar (0-3)
}

// NewLiveDisplay creates a new live display
func NewLiveDisplay(track *parser.Track) *LiveDisplay {
	// Calculate time per beat based on tempo
	beatsPerSecond := float64(track.Info.Tempo) / 60.0
	timePerBeat := time.Duration(float64(time.Second) / beatsPerSecond)

	// Get strum pattern
	strumPattern := getStrumPattern(track.Rhythm)

	// Process chords into bars
	bars := processChordsIntoBars(track)

	// Initialize scale based on track style
	scale := theory.GetScaleForStyle(track.Info.Key, track.Info.Style, "")

	// Create fretboard display (15 frets, compact mode for now)
	fretboard := NewFretboardDisplay(scale, 15)
	fretboard.SetCompactMode(true) // Use compact mode to fit alongside chord display

	// Create chord chart
	chordChart := NewChordChart()

	// Get initial chord
	initialChord := ""
	if len(bars) > 0 && len(bars[0].Chords) > 0 {
		initialChord = bars[0].Chords[0].Symbol
	}

	return &LiveDisplay{
		track:         track,
		chords:        track.Progression.GetChords(),
		bars:          bars,
		tempo:         track.Info.Tempo,
		timePerBeat:   timePerBeat,
		stopChan:      make(chan bool),
		strumPattern:  strumPattern,
		barsPerLine:   2, // 2 bars per line for karaoke style (more readable)
		displayLines:  0,
		fretboard:     fretboard,
		currentScale:  scale,
		showFretboard: true,
		chordChart:    chordChart,
		currentChord:  initialChord,
	}
}

// processChordsIntoBars converts chord progression into bar structure
func processChordsIntoBars(track *parser.Track) []Bar {
	chords := track.Progression.GetChords()
	var bars []Bar

	currentBar := Bar{Chords: []BarChord{}, Lyrics: ""}
	currentBeatInBar := 0

	for _, chord := range chords {
		beatsForChord := int(chord.Bars * 4)

		// Handle chord that fits in current bar
		for beatsForChord > 0 {
			beatsAvailable := 4 - currentBeatInBar
			beatsToUse := beatsForChord
			if beatsToUse > beatsAvailable {
				beatsToUse = beatsAvailable
			}

			currentBar.Chords = append(currentBar.Chords, BarChord{
				Symbol:    chord.Symbol,
				Beats:     beatsToUse,
				StartBeat: currentBeatInBar,
			})

			currentBeatInBar += beatsToUse
			beatsForChord -= beatsToUse

			// If bar is full, start a new one
			if currentBeatInBar >= 4 {
				bars = append(bars, currentBar)
				currentBar = Bar{Chords: []BarChord{}, Lyrics: ""}
				currentBeatInBar = 0
			}
		}
	}

	// Add any remaining partial bar
	if len(currentBar.Chords) > 0 {
		bars = append(bars, currentBar)
	}

	// Add lyrics to bars
	if track.Lyrics != nil {
		for i, lyric := range track.Lyrics {
			if i < len(bars) {
				bars[i].Lyrics = lyric
			}
		}
	}

	return bars
}

// getStrumPattern returns the strum pattern string
func getStrumPattern(rhythm *parser.Rhythm) string {
	if rhythm == nil {
		return "↓ . ↓ . ↓ . ↓ ." // Default quarter notes
	}

	if rhythm.Pattern != "" {
		return convertPatternToDisplay(rhythm.Pattern)
	}

	switch rhythm.Style {
	case "whole":
		return "↓ - - - - - - -"
	case "half":
		return "↓ - - - ↓ - - -"
	case "quarter":
		return "↓ - ↓ - ↓ - ↓ -"
	case "eighth":
		return "↓ . ↑ . ↓ . ↑ ."
	case "strum_down":
		return "↓ . . ↓ . . ↓ ."
	case "strum_up_down":
		return "↓ . ↑ . ↓ . ↑ ."
	case "folk":
		return "↓ . ↓ ↑ ↓ . ↓ ↑"
	case "shuffle_strum":
		return ". ↓ . ↓ . ↓ . ↑"
	case "travis", "fingerpick":
		return "↓ . ↑ . ↓ . ↑ ."
	case "fingerpick_slow":
		return "↓ . . . ↓ . . ."
	case "arpeggio_up", "arpeggio_down":
		return "↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓"
	case "stride":
		return ". . ↓ . . . ↓ ."
	case "ragtime":
		return ". ↓ ↓ . . ↓ ↓ ."
	case "sixteenth":
		return "↓.↑.↓.↑.↓.↑.↓.↑."
	case "funk_16th":
		return "↓.x.↑x↓.x.↑.↓x↑."
	case "funk_muted":
		return "x.↓.x.↑.x.↓.x.↑."
	default:
		return "↓ . ↑ . ↓ . ↑ ."
	}
}

// convertPatternToDisplay converts D/U/x/. pattern to display symbols
func convertPatternToDisplay(pattern string) string {
	var result []string
	for _, c := range pattern {
		switch c {
		case 'D':
			result = append(result, "↓")
		case 'd':
			result = append(result, "↓")
		case 'U':
			result = append(result, "↑")
		case 'u':
			result = append(result, "↑")
		case 'x':
			result = append(result, "x")
		case '.':
			result = append(result, ".")
		case '-':
			result = append(result, "-")
		}
	}
	// Pad to 8 if needed
	for len(result) < 8 {
		result = append(result, ".")
	}
	return strings.Join(result, " ")
}

// Start begins the live display
func (ld *LiveDisplay) Start() {
	ld.startTime = time.Now()
	go ld.run()
}

// Stop stops the live display
func (ld *LiveDisplay) Stop() {
	ld.stopChan <- true
}

// run is the main display loop
func (ld *LiveDisplay) run() {
	ticker := time.NewTicker(50 * time.Millisecond) // Update 20 times per second for smooth strum display
	defer ticker.Stop()

	// Initial render to calculate display lines
	ld.renderFull()

	for {
		select {
		case <-ld.stopChan:
			return
		case <-ticker.C:
			ld.render()
		}
	}
}

// Number of bar rows to show at once (scrolling window)
const visibleRows = 4

// getUniqueChords returns all unique chord symbols from the song
func (ld *LiveDisplay) getUniqueChords() []string {
	seen := make(map[string]bool)
	var unique []string
	for _, bar := range ld.bars {
		for _, bc := range bar.Chords {
			// Strip slash chord bass note for chart lookup
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

// renderAllChordCharts renders chord charts for all unique chords in the song
// For fingerpicking styles, also includes the picking pattern tablature
func (ld *LiveDisplay) renderAllChordCharts() []string {
	var lines []string

	// Add fingerpicking pattern if applicable
	if ld.isFingerPickingStyle() {
		lines = append(lines, " Picking Pattern:")
		lines = append(lines, ld.getFingerPickingPattern()...)
		lines = append(lines, "")
	}

	if ld.chordChart == nil {
		return lines
	}

	uniqueChords := ld.getUniqueChords()

	for _, chord := range uniqueChords {
		voicings := ld.chordChart.GetVoicings(chord)
		if len(voicings) == 0 {
			lines = append(lines, fmt.Sprintf(" %s: [no chart]", chord))
			lines = append(lines, "")
			continue
		}

		// Show all voicings for this chord
		for _, v := range voicings {
			chordLines := ld.chordChart.RenderSingleChord(v)
			lines = append(lines, chordLines...)
			lines = append(lines, "") // spacer between voicings
		}
	}

	return lines
}

// renderFull renders the initial display
func (ld *LiveDisplay) renderFull() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Render header with scale info
	fmt.Printf("  %s", ld.track.Info.Title)
	titlePad := 50 - len(ld.track.Info.Title)
	if titlePad < 1 {
		titlePad = 1
	}
	fmt.Printf("%s%s | %d BPM",
		strings.Repeat(" ", titlePad),
		ld.track.Info.Key,
		ld.track.Info.Tempo)

	// Add scale name to header if fretboard is shown
	if ld.showFretboard && ld.currentScale != nil {
		fmt.Printf("  │  Scale: %s", ld.currentScale.Name)
	}
	fmt.Println()

	fmt.Println("  " + strings.Repeat("═", 66))
	fmt.Println()

	// Render first few rows with fretboard and chord chart on right
	fretLines := []string{}
	if ld.showFretboard && ld.fretboard != nil {
		fretLines = ld.fretboard.Render()
	}

	// Get chord charts for ALL unique chords in the song
	chordLines := ld.renderAllChordCharts()

	// Combine fretboard and chord chart
	rightPanelLines := append(fretLines, chordLines...)

	for row := 0; row < visibleRows; row++ {
		lineStart := row * ld.barsPerLine
		if lineStart < len(ld.bars) {
			ld.renderBarLineWithFretboard(lineStart, -1, -1, -1, rightPanelLines, row)
		} else {
			// Empty rows
			for i := 0; i < 5; i++ {
				fmt.Println()
			}
		}
	}

	// Progress bar
	fmt.Println()
	fmt.Printf("  %s  0%%\n", strings.Repeat("░", 50))
}

// render updates the display with scrolling
func (ld *LiveDisplay) render() {
	elapsed := time.Since(ld.startTime)

	// Calculate current position
	totalBeats := int(elapsed / ld.timePerBeat)
	currentBeat := totalBeats % 4          // 0-3
	currentBar := totalBeats / 4           // Which bar we're in

	// Calculate strum position (8 strums per bar for 8th notes)
	timePerStrum := ld.timePerBeat / 2
	totalStrums := int(elapsed / timePerStrum)
	currentStrum := totalStrums % 8 // 0-7

	// Calculate which row the current bar is in
	currentRow := currentBar / ld.barsPerLine

	// Calculate visible window (keep current bar in the second row if possible)
	startRow := currentRow - 1
	if startRow < 0 {
		startRow = 0
	}
	// Don't scroll past the end
	totalRows := (len(ld.bars) + ld.barsPerLine - 1) / ld.barsPerLine
	if startRow + visibleRows > totalRows {
		startRow = totalRows - visibleRows
		if startRow < 0 {
			startRow = 0
		}
	}

	// Update scale and current chord if changed
	if currentBar < len(ld.bars) && len(ld.bars[currentBar].Chords) > 0 {
		newChord := ld.bars[currentBar].Chords[0].Symbol
		if newChord != ld.currentChord {
			ld.currentChord = newChord
		}
		if strings.Contains(strings.ToLower(ld.track.Info.Style), "jazz") {
			newScale := theory.GetScaleForStyle(ld.track.Info.Key, ld.track.Info.Style, newChord)
			if newScale.Name != ld.currentScale.Name {
				ld.currentScale = newScale
				ld.fretboard.SetScale(newScale)
			}
		}
	}

	// Get fretboard lines
	fretLines := []string{}
	if ld.showFretboard && ld.fretboard != nil {
		fretLines = ld.fretboard.Render()
	}

	// Get chord charts for ALL unique chords in the song
	chordLines := ld.renderAllChordCharts()

	// Combine fretboard and chord chart
	rightPanelLines := append(fretLines, chordLines...)

	// Move cursor to start
	fmt.Print("\033[H")

	// Skip header (3 lines: title, separator, empty line)
	fmt.Print("\033[3B")

	// Render visible bar rows with fretboard and chord chart
	for row := 0; row < visibleRows; row++ {
		lineStart := (startRow + row) * ld.barsPerLine
		if lineStart < len(ld.bars) {
			ld.renderBarLineWithFretboard(lineStart, currentBar, currentBeat, currentStrum, rightPanelLines, row)
		} else {
			// Empty rows (clear them)
			for i := 0; i < 5; i++ {
				fmt.Print("\033[2K\n") // Clear line
			}
		}
	}

	// Progress bar
	fmt.Println()
	var progress float64
	if len(ld.bars) > 0 {
		progress = float64(currentBar) / float64(len(ld.bars))
	}
	if progress < 0 {
		progress = 0
	} else if progress > 1.0 {
		progress = 1.0
	}
	filledWidth := int(progress * 50)
	progressBar := strings.Repeat("▓", filledWidth) + strings.Repeat("░", 50-filledWidth)
	fmt.Printf("  %s  %d%% (bar %d/%d)\033[K\n", progressBar, int(progress*100), currentBar+1, len(ld.bars))
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
)

// visibleLength returns the visible length of a string, ignoring ANSI codes
func visibleLength(s string) int {
	// Remove ANSI escape sequences
	inEscape := false
	visible := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		visible++
	}
	return visible
}

// padToWidth pads a string to the given visible width
func padToWidth(s string, width int) string {
	visLen := visibleLength(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// renderBarLine renders a line of bars (2 bars per line)
func (ld *LiveDisplay) renderBarLine(startBar int, currentBar int, currentBeat int, currentStrum int) {
	endBar := startBar + ld.barsPerLine
	if endBar > len(ld.bars) {
		endBar = len(ld.bars)
	}

	barWidth := 31 // Width for each bar content

	// Line 1: Chord names
	fmt.Print("  ")
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			chordStr := ld.formatBarChords(ld.bars[i], barWidth)
			if i == currentBar {
				fmt.Printf("%s%s%s%s  ", colorBold, colorCyan, chordStr, colorReset)
			} else {
				fmt.Printf("%s  ", chordStr)
			}
		}
	}
	fmt.Println()

	// Line 2: Lyrics (always render to maintain consistent row height)
	fmt.Print("  ")
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			lyrics := ld.bars[i].Lyrics
			if len(lyrics) > barWidth {
				lyrics = lyrics[:barWidth]
			}
			padded := fmt.Sprintf("%-*s", barWidth, lyrics)
			if i == currentBar && lyrics != "" {
				fmt.Printf("%s%s%s%s  ", colorBold, colorYellow, padded, colorReset)
			} else {
				fmt.Printf("%s  ", padded)
			}
		}
	}
	fmt.Println()

	// Line 3: Strum pattern
	fmt.Print("  ")
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			strumDisplay := ld.formatStrumPattern(i == currentBar, currentStrum, barWidth)
			if i == currentBar {
				fmt.Printf("%s%s%s  ", colorGreen, strumDisplay, colorReset)
			} else {
				fmt.Printf("%s%s%s  ", colorDim, strumDisplay, colorReset)
			}
		}
	}
	fmt.Println()

	// Line 4: Beat numbers
	fmt.Print("  ")
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			var beatDisplay string
			if ld.isSixteenthNoteStyle() {
				beatDisplay = ld.formatBeatNumbers16th(i == currentBar, currentBeat, barWidth)
			} else {
				beatDisplay = ld.formatBeatNumbers(i == currentBar, currentBeat, barWidth)
			}
			if i == currentBar {
				fmt.Printf("%s%s%s  ", colorGreen, beatDisplay, colorReset)
			} else {
				fmt.Printf("%s%s%s  ", colorDim, beatDisplay, colorReset)
			}
		}
	}
	fmt.Println()

	// Line 5: Separator
	fmt.Print("  ")
	fmt.Println(strings.Repeat("─", (barWidth+2)*ld.barsPerLine))
}

// formatBarChords formats chord names for display
func (ld *LiveDisplay) formatBarChords(bar Bar, width int) string {
	if len(bar.Chords) == 0 {
		return strings.Repeat(" ", width)
	}

	if len(bar.Chords) == 1 {
		// Single chord - center it
		chord := bar.Chords[0].Symbol
		padding := (width - len(chord)) / 2
		return fmt.Sprintf("%s%s%s",
			strings.Repeat(" ", padding),
			chord,
			strings.Repeat(" ", width-padding-len(chord)))
	}

	// Multiple chords - divide the space
	parts := make([]string, len(bar.Chords))
	partWidth := width / len(bar.Chords)

	for i, bc := range bar.Chords {
		chord := bc.Symbol
		if len(chord) > partWidth-1 {
			chord = chord[:partWidth-1]
		}
		padding := (partWidth - len(chord)) / 2
		parts[i] = fmt.Sprintf("%s%s%s",
			strings.Repeat(" ", padding),
			chord,
			strings.Repeat(" ", partWidth-padding-len(chord)))
	}

	result := strings.Join(parts, "│")
	// Pad or trim to exact width
	if len(result) < width {
		result += strings.Repeat(" ", width-len(result))
	} else if len(result) > width {
		result = result[:width]
	}
	return result
}

// formatStrumPattern formats the strum pattern with current position highlighted
func (ld *LiveDisplay) formatStrumPattern(isCurrentBar bool, currentStrum int, width int) string {
	parts := strings.Split(ld.strumPattern, " ")
	if len(parts) < 8 {
		// Pad to 8
		for len(parts) < 8 {
			parts = append(parts, ".")
		}
	}

	var result []string
	for i, p := range parts[:8] {
		if isCurrentBar {
			if i == currentStrum {
				result = append(result, "█") // Current position
			} else if i > currentStrum {
				result = append(result, "░") // Upcoming
			} else {
				result = append(result, p) // Passed
			}
		} else {
			result = append(result, p)
		}
	}

	// Format with spacing: 8 symbols with 3 spaces between = 30 chars
	// Add 1 space prefix to help center in width 31
	display := " " + strings.Join(result, "   ")

	// Pad to exact width (use rune count for Unicode)
	runeCount := utf8.RuneCountInString(display)
	if runeCount < width {
		display = display + strings.Repeat(" ", width-runeCount)
	}

	return display
}

// formatBeatNumbers formats beat numbers with current beat highlighted
func (ld *LiveDisplay) formatBeatNumbers(isCurrentBar bool, currentBeat int, width int) string {
	beats := []string{"1", "2", "3", "4"}

	var result []string
	for i, b := range beats {
		if isCurrentBar {
			if i == currentBeat {
				result = append(result, "●") // Current beat
			} else if i == 0 {
				result = append(result, "◉") // Beat 1 (downbeat)
			} else if i > currentBeat {
				result = append(result, "○") // Upcoming
			} else {
				result = append(result, b) // Passed
			}
		} else {
			result = append(result, b)
		}
	}

	// Format with spacing to match strum pattern exactly
	// Strum: " x   x   x   x   x   x   x   x" (1 space prefix + 8 symbols with 3 spaces)
	// Beats align with strums 0, 2, 4, 6 at positions 1, 9, 17, 25
	// So: 1 space prefix, then 7 spaces between each beat number
	display := fmt.Sprintf(" %s       %s       %s       %s", result[0], result[1], result[2], result[3])

	// Pad to exact width (use rune count for Unicode symbols like ●)
	runeCount := utf8.RuneCountInString(display)
	if runeCount < width {
		display = display + strings.Repeat(" ", width-runeCount)
	}

	return display
}

// formatBeatNumbers16th formats beat numbers for 16th note patterns
func (ld *LiveDisplay) formatBeatNumbers16th(isCurrentBar bool, currentBeat int, width int) string {
	// 16th note subdivisions: 1 e + a 2 e + a 3 e + a 4 e + a
	beats := []string{"1", "e", "+", "a", "2", "e", "+", "a", "3", "e", "+", "a", "4", "e", "+", "a"}

	var result []string
	for i, b := range beats {
		beatNum := i / 4 // Which quarter note beat (0-3)
		if isCurrentBar {
			if beatNum == currentBeat && i%4 == 0 {
				result = append(result, "●") // Current beat downbeat
			} else if i == 0 && beatNum != currentBeat {
				result = append(result, "◉") // Beat 1 (downbeat) when not current
			} else {
				result = append(result, b)
			}
		} else {
			result = append(result, b)
		}
	}

	// Format: "1 e + a 2 e + a 3 e + a 4 e + a"
	display := " " + strings.Join(result, " ")

	runeCount := utf8.RuneCountInString(display)
	if runeCount < width {
		display = display + strings.Repeat(" ", width-runeCount)
	}

	return display
}

// isSixteenthNoteStyle returns true if the rhythm style uses 16th notes
func (ld *LiveDisplay) isSixteenthNoteStyle() bool {
	if ld.track.Rhythm == nil {
		return false
	}
	style := ld.track.Rhythm.Style
	return style == "sixteenth" || style == "funk_16th" || style == "funk_muted"
}

// isFingerPickingStyle returns true if the rhythm style is fingerpicking
func (ld *LiveDisplay) isFingerPickingStyle() bool {
	if ld.track.Rhythm == nil {
		return false
	}
	style := ld.track.Rhythm.Style
	return style == "fingerpick" || style == "fingerpick_slow" || style == "travis" ||
		style == "arpeggio_up" || style == "arpeggio_down"
}

// getFingerPickingPattern returns tablature lines for fingerpicking
func (ld *LiveDisplay) getFingerPickingPattern() []string {
	if ld.track.Rhythm == nil {
		return []string{}
	}

	style := ld.track.Rhythm.Style
	switch style {
	case "fingerpick_slow":
		// Slow fingerpick: p i m a pattern, sparse
		return []string{
			"e|----0-------0---|",
			"B|------0-------0-|",
			"G|--0-------0-----|",
			"D|----------------|",
			"A|----------------|",
			"E|0-------0-------|",
		}
	case "fingerpick":
		// Standard fingerpick: continuous 16th pattern
		return []string{
			"e|----0---0---0---|",
			"B|------0---0---0-|",
			"G|--0---0---0---0-|",
			"D|----------------|",
			"A|----------------|",
			"E|0---0---0---0---|",
		}
	case "travis":
		// Travis picking: alternating bass with melody
		return []string{
			"e|------0---0-----|",
			"B|----0---0---0---|",
			"G|--0-------0-----|",
			"D|----------------|",
			"A|----0-------0---|",
			"E|0-------0-------|",
		}
	case "arpeggio_up":
		// Ascending arpeggio
		return []string{
			"e|------0---------|",
			"B|----0-----------|",
			"G|--0-------------|",
			"D|0---------------|",
			"A|----------------|",
			"E|----------------|",
		}
	case "arpeggio_down":
		// Descending arpeggio
		return []string{
			"e|0---------------|",
			"B|--0-------------|",
			"G|----0-----------|",
			"D|------0---------|",
			"A|----------------|",
			"E|----------------|",
		}
	default:
		return []string{}
	}
}

// renderBarLineWithFretboard renders a bar line with three columns:
// Left: chord/beat display, Middle: scale fretboard, Right: chord charts
func (ld *LiveDisplay) renderBarLineWithFretboard(startBar int, currentBar int, currentBeat int, currentStrum int, rightPanelLines []string, rowIndex int) {
	endBar := startBar + ld.barsPerLine
	if endBar > len(ld.bars) {
		endBar = len(ld.bars)
	}

	barWidth := 32 // Full width for each bar (fits longer lyrics)
	leftColWidth := (barWidth + 2) * ld.barsPerLine
	middleColWidth := 42 // Width for scale fretboard

	// Calculate line offsets for each column (5 lines per display row)
	fretStartLine := rowIndex * 5

	// Get scale lines (middle column)
	scaleLines := []string{}
	if ld.showFretboard && ld.fretboard != nil {
		scaleLines = ld.fretboard.Render()
	}

	// Get chord chart lines (right column) - already in rightPanelLines after scale
	chordChartStart := len(scaleLines)
	chordLines := []string{}
	if chordChartStart < len(rightPanelLines) {
		chordLines = rightPanelLines[chordChartStart:]
	}

	// Helper to get line from slice safely
	getLine := func(lines []string, idx int) string {
		if idx >= 0 && idx < len(lines) {
			return lines[idx]
		}
		return ""
	}

	// Line 1: Chord names
	leftContent := "  "
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			chordStr := ld.formatBarChords(ld.bars[i], barWidth)
			if i == currentBar {
				leftContent += fmt.Sprintf("%s%s%s%s  ", colorBold, colorCyan, chordStr, colorReset)
			} else {
				leftContent += fmt.Sprintf("%s  ", chordStr)
			}
		}
	}
	fmt.Print(padToWidth(leftContent, leftColWidth+2))
	fmt.Print(" │ ")
	fmt.Print(padToWidth(getLine(scaleLines, fretStartLine), middleColWidth))
	fmt.Print(" │ ")
	fmt.Print(getLine(chordLines, fretStartLine))
	fmt.Print("\033[K\n")

	// Line 2: Lyrics
	leftContent = "  "
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			lyrics := ld.bars[i].Lyrics
			if len(lyrics) > barWidth {
				lyrics = lyrics[:barWidth]
			}
			padded := fmt.Sprintf("%-*s", barWidth, lyrics)
			if i == currentBar && lyrics != "" {
				leftContent += fmt.Sprintf("%s%s%s%s  ", colorBold, colorYellow, padded, colorReset)
			} else {
				leftContent += fmt.Sprintf("%s  ", padded)
			}
		}
	}
	fmt.Print(padToWidth(leftContent, leftColWidth+2))
	fmt.Print(" │ ")
	fmt.Print(padToWidth(getLine(scaleLines, fretStartLine+1), middleColWidth))
	fmt.Print(" │ ")
	fmt.Print(getLine(chordLines, fretStartLine+1))
	fmt.Print("\033[K\n")

	// Line 3: Strum pattern
	leftContent = "  "
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			strumDisplay := ld.formatStrumPattern(i == currentBar, currentStrum, barWidth)
			if i == currentBar {
				leftContent += fmt.Sprintf("%s%s%s  ", colorGreen, strumDisplay, colorReset)
			} else {
				leftContent += fmt.Sprintf("%s%s%s  ", colorDim, strumDisplay, colorReset)
			}
		}
	}
	fmt.Print(padToWidth(leftContent, leftColWidth+2))
	fmt.Print(" │ ")
	fmt.Print(padToWidth(getLine(scaleLines, fretStartLine+2), middleColWidth))
	fmt.Print(" │ ")
	fmt.Print(getLine(chordLines, fretStartLine+2))
	fmt.Print("\033[K\n")

	// Line 4: Beat numbers
	leftContent = "  "
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			var beatDisplay string
			if ld.isSixteenthNoteStyle() {
				beatDisplay = ld.formatBeatNumbers16th(i == currentBar, currentBeat, barWidth)
			} else {
				beatDisplay = ld.formatBeatNumbers(i == currentBar, currentBeat, barWidth)
			}
			if i == currentBar {
				leftContent += fmt.Sprintf("%s%s%s  ", colorGreen, beatDisplay, colorReset)
			} else {
				leftContent += fmt.Sprintf("%s%s%s  ", colorDim, beatDisplay, colorReset)
			}
		}
	}
	fmt.Print(padToWidth(leftContent, leftColWidth+2))
	fmt.Print(" │ ")
	fmt.Print(padToWidth(getLine(scaleLines, fretStartLine+3), middleColWidth))
	fmt.Print(" │ ")
	fmt.Print(getLine(chordLines, fretStartLine+3))
	fmt.Print("\033[K\n")

	// Line 5: Separator
	leftContent = "  " + strings.Repeat("─", (barWidth+2)*ld.barsPerLine)
	fmt.Print(padToWidth(leftContent, leftColWidth+2))
	fmt.Print(" │ ")
	fmt.Print(padToWidth(getLine(scaleLines, fretStartLine+4), middleColWidth))
	fmt.Print(" │ ")
	fmt.Print(getLine(chordLines, fretStartLine+4))
	fmt.Print("\033[K\n")
}

// formatBarChordsCompact formats chord names in compact form
func (ld *LiveDisplay) formatBarChordsCompact(bar Bar, width int) string {
	if len(bar.Chords) == 0 {
		return strings.Repeat(" ", width)
	}

	if len(bar.Chords) == 1 {
		chord := bar.Chords[0].Symbol
		padding := (width - len(chord)) / 2
		if padding < 0 {
			padding = 0
		}
		result := fmt.Sprintf("%s%s", strings.Repeat(" ", padding), chord)
		if len(result) < width {
			result += strings.Repeat(" ", width-len(result))
		}
		return result
	}

	// Multiple chords
	parts := []string{}
	for _, bc := range bar.Chords {
		parts = append(parts, bc.Symbol)
	}
	result := strings.Join(parts, "|")
	if len(result) < width {
		padding := (width - len(result)) / 2
		result = strings.Repeat(" ", padding) + result + strings.Repeat(" ", width-padding-len(result))
	}
	return result
}

// formatBeatDots formats beat position as dots
func (ld *LiveDisplay) formatBeatDots(isCurrentBar bool, currentBeat int, width int) string {
	var result []string
	for i := 0; i < 4; i++ {
		if isCurrentBar && i == currentBeat {
			result = append(result, "█")
		} else if isCurrentBar && i < currentBeat {
			result = append(result, "·")
		} else {
			result = append(result, ".")
		}
	}
	display := strings.Join(result, "   ")
	if len(display) < width {
		display = display + strings.Repeat(" ", width-len(display))
	}
	return display
}

// formatBeatNumbersCompact formats beat numbers compactly
func (ld *LiveDisplay) formatBeatNumbersCompact(isCurrentBar bool, currentBeat int, width int) string {
	beats := []string{"1", "2", "3", "4"}
	var result []string
	for i, b := range beats {
		if isCurrentBar && i == currentBeat {
			result = append(result, "●")
		} else {
			result = append(result, b)
		}
	}
	display := strings.Join(result, "   ")
	runeCount := utf8.RuneCountInString(display)
	if runeCount < width {
		display = display + strings.Repeat(" ", width-runeCount)
	}
	return display
}
