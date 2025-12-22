package display

import (
	"fmt"
	"strings"
	"time"

	"backing-tracks/parser"
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

	return &LiveDisplay{
		track:        track,
		chords:       track.Progression.GetChords(),
		bars:         bars,
		tempo:        track.Info.Tempo,
		timePerBeat:  timePerBeat,
		stopChan:     make(chan bool),
		strumPattern: strumPattern,
		barsPerLine:  2, // 2 bars per line for karaoke style (more readable)
		displayLines: 0,
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

// renderFull renders the complete display (initial)
func (ld *LiveDisplay) renderFull() {
	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Render header
	fmt.Printf("  %s", ld.track.Info.Title)
	fmt.Printf("%s%s | %d BPM\n",
		strings.Repeat(" ", 50-len(ld.track.Info.Title)),
		ld.track.Info.Key,
		ld.track.Info.Tempo)
	fmt.Println("  " + strings.Repeat("═", 66))
	fmt.Println()

	ld.displayLines = 3 // Header lines

	// Render all bars
	for lineStart := 0; lineStart < len(ld.bars); lineStart += ld.barsPerLine {
		ld.renderBarLine(lineStart, -1, -1, -1)
		ld.displayLines += 5 // Each bar line takes 5 lines
	}

	// Progress bar
	fmt.Println()
	fmt.Printf("  %s  0%%\n", strings.Repeat("░", 50))
	ld.displayLines += 2
}

// render updates just the dynamic parts
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

	// Move cursor to start
	fmt.Print("\033[H")

	// Skip header
	fmt.Print("\033[3B")

	// Render all bar lines with current position
	for lineStart := 0; lineStart < len(ld.bars); lineStart += ld.barsPerLine {
		ld.renderBarLine(lineStart, currentBar, currentBeat, currentStrum)
	}

	// Progress bar
	fmt.Println()
	progress := float64(currentBar) / float64(len(ld.bars))
	if progress > 1.0 {
		progress = 1.0
	}
	filledWidth := int(progress * 50)
	progressBar := strings.Repeat("▓", filledWidth) + strings.Repeat("░", 50-filledWidth)
	fmt.Printf("  %s  %d%%\n", progressBar, int(progress*100))
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

	// Check if this line contains the current bar
	isCurrentLine := currentBar >= startBar && currentBar < endBar

	// Line 1: Chord names with bar numbers
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
	fmt.Printf("%d-%d\n", startBar+1, endBar)

	// Line 2: Lyrics
	fmt.Print("  ")
	for i := startBar; i < endBar; i++ {
		if i < len(ld.bars) {
			lyrics := ld.bars[i].Lyrics
			if len(lyrics) > barWidth {
				lyrics = lyrics[:barWidth]
			}
			padded := fmt.Sprintf("%-*s", barWidth, lyrics)
			if i == currentBar {
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

	// Line 4: Beat numbers with current line marker
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
	if isCurrentLine {
		fmt.Printf("%s◄───%s", colorCyan, colorReset)
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

	// Format with spacing
	display := strings.Join(result, "   ")

	// Center in width
	if len(display) < width {
		padding := (width - len(display)) / 2
		display = strings.Repeat(" ", padding) + display + strings.Repeat(" ", width-padding-len(display))
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

	// Format with spacing to match strum pattern (8 positions = 4 beats * 2)
	display := fmt.Sprintf("%s       %s       %s       %s", result[0], result[1], result[2], result[3])

	// Center in width
	if len(display) < width {
		padding := (width - len(display)) / 2
		display = strings.Repeat(" ", padding) + display + strings.Repeat(" ", width-padding-len(display))
	}

	return display
}
