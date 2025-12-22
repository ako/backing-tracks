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

	// Get chord chart lines for initial chord
	chordLines := []string{}
	if ld.chordChart != nil && ld.currentChord != "" {
		chordLines = ld.chordChart.RenderHorizontal(ld.currentChord)
	}

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

	// Get chord chart lines for current chord
	chordLines := []string{}
	if ld.chordChart != nil && ld.currentChord != "" {
		chordLines = ld.chordChart.RenderHorizontal(ld.currentChord)
	}

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
			beatDisplay := ld.formatBeatNumbers(i == currentBar, currentBeat, barWidth)
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

// renderBarLineWithFretboard renders a bar line with fretboard on the right
func (ld *LiveDisplay) renderBarLineWithFretboard(startBar int, currentBar int, currentBeat int, currentStrum int, fretLines []string, rowIndex int) {
	endBar := startBar + ld.barsPerLine
	if endBar > len(ld.bars) {
		endBar = len(ld.bars)
	}

	barWidth := 31 // Width for each bar content
	leftWidth := (barWidth + 2) * ld.barsPerLine + 2 // Total left side width

	// Calculate which fretboard lines to show for this row
	// Each bar row is 5 lines, fretboard has about 10 lines
	// Show fretboard lines spread across the rows
	fretStartLine := rowIndex * 5
	if fretStartLine >= len(fretLines) {
		fretStartLine = 0
	}

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
	// Add fretboard line on right
	if ld.showFretboard && fretStartLine < len(fretLines) {
		padNeeded := leftWidth - (barWidth+2)*(endBar-startBar) - 2
		if padNeeded > 0 {
			fmt.Print(strings.Repeat(" ", padNeeded))
		}
		fmt.Printf("    │  %s", fretLines[fretStartLine])
	}
	fmt.Print("\033[K\n")

	// Line 2: Lyrics
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
	// Add fretboard line on right
	if ld.showFretboard && fretStartLine+1 < len(fretLines) {
		padNeeded := leftWidth - (barWidth+2)*(endBar-startBar) - 2
		if padNeeded > 0 {
			fmt.Print(strings.Repeat(" ", padNeeded))
		}
		fmt.Printf("    │  %s", fretLines[fretStartLine+1])
	}
	fmt.Print("\033[K\n")

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
	// Add fretboard line on right
	if ld.showFretboard && fretStartLine+2 < len(fretLines) {
		padNeeded := leftWidth - (barWidth+2)*(endBar-startBar) - 2
		if padNeeded > 0 {
			fmt.Print(strings.Repeat(" ", padNeeded))
		}
		fmt.Printf("    │  %s", fretLines[fretStartLine+2])
	}
	fmt.Print("\033[K\n")

	// Line 4: Beat numbers
	fmt.Print("  ")
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			beatDisplay := ld.formatBeatNumbers(i == currentBar, currentBeat, barWidth)
			if i == currentBar {
				fmt.Printf("%s%s%s  ", colorGreen, beatDisplay, colorReset)
			} else {
				fmt.Printf("%s%s%s  ", colorDim, beatDisplay, colorReset)
			}
		}
	}
	// Add fretboard line on right
	if ld.showFretboard && fretStartLine+3 < len(fretLines) {
		padNeeded := leftWidth - (barWidth+2)*(endBar-startBar) - 2
		if padNeeded > 0 {
			fmt.Print(strings.Repeat(" ", padNeeded))
		}
		fmt.Printf("    │  %s", fretLines[fretStartLine+3])
	}
	fmt.Print("\033[K\n")

	// Line 5: Separator
	fmt.Print("  ")
	fmt.Print(strings.Repeat("─", (barWidth+2)*ld.barsPerLine))
	// Add fretboard line on right
	if ld.showFretboard && fretStartLine+4 < len(fretLines) {
		fmt.Printf("    │  %s", fretLines[fretStartLine+4])
	}
	fmt.Print("\033[K\n")
}
