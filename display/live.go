package display

import (
	"fmt"
	"strings"
	"time"

	"backing-tracks/parser"
)

// LiveDisplay shows real-time playback information
type LiveDisplay struct {
	track       *parser.Track
	chords      []parser.Chord
	tempo       int
	timePerBeat time.Duration
	startTime   time.Time
	stopChan    chan bool
}

// NewLiveDisplay creates a new live display
func NewLiveDisplay(track *parser.Track) *LiveDisplay {
	// Calculate time per beat based on tempo
	// tempo is in BPM (beats per minute)
	beatsPerSecond := float64(track.Info.Tempo) / 60.0
	timePerBeat := time.Duration(float64(time.Second) / beatsPerSecond)

	return &LiveDisplay{
		track:       track,
		chords:      track.Progression.GetChords(),
		tempo:       track.Info.Tempo,
		timePerBeat: timePerBeat,
		stopChan:    make(chan bool),
	}
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
	ticker := time.NewTicker(100 * time.Millisecond) // Update 10 times per second
	defer ticker.Stop()

	// Clear space for display
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	for {
		select {
		case <-ld.stopChan:
			return
		case <-ticker.C:
			ld.render()
		}
	}
}

// render updates the display
func (ld *LiveDisplay) render() {
	elapsed := time.Since(ld.startTime)

	// Calculate current position
	totalBeats := int(elapsed / ld.timePerBeat)
	currentBeat := (totalBeats % 4) + 1 // 1-4 for 4/4 time
	totalBars := totalBeats / 4

	// Find current chord
	currentChordIndex := -1
	beatsIntoProgression := totalBeats

	for i, chord := range ld.chords {
		// Calculate beats in this chord (4 beats per bar)
		// Support fractional bars (0.5 bar = 2 beats, 1 bar = 4 beats)
		beatsInChord := int(chord.Bars * 4)
		if beatsIntoProgression < beatsInChord {
			currentChordIndex = i
			break
		}
		beatsIntoProgression -= beatsInChord
	}

	// If we've finished, show the last chord
	if currentChordIndex == -1 && len(ld.chords) > 0 {
		currentChordIndex = len(ld.chords) - 1
	}

	// Move cursor up 5 lines to overwrite previous output
	fmt.Print("\033[5A")

	// Clear lines
	for i := 0; i < 5; i++ {
		fmt.Print("\033[2K") // Clear line
		fmt.Println()
	}

	// Move cursor back up
	fmt.Print("\033[5A")

	// Render current state
	if currentChordIndex >= 0 && currentChordIndex < len(ld.chords) {
		currentChord := ld.chords[currentChordIndex]

		// Show current chord (large and prominent)
		fmt.Printf("┌─────────────────────────────────────┐\n")
		fmt.Printf("│  Current Chord: %-20s│\n", padChord(currentChord.Symbol))
		fmt.Printf("└─────────────────────────────────────┘\n")

		// Show beat indicator (visual metronome)
		beatDisplay := renderBeatIndicator(currentBeat)
		fmt.Printf("Beat: %s  [Bar %d, Beat %d]\n", beatDisplay, totalBars+1, currentBeat)

		// Show progress through progression
		progressBar := renderProgressBar(currentChordIndex, len(ld.chords))
		fmt.Printf("Progress: %s %d/%d\n", progressBar, currentChordIndex+1, len(ld.chords))
	} else {
		fmt.Println("Playback starting...")
		fmt.Println()
		fmt.Println()
		fmt.Println()
	}
}

// padChord pads a chord symbol for display
func padChord(chord string) string {
	// Make chord display large
	return fmt.Sprintf("%-8s", chord)
}

// renderBeatIndicator creates a visual metronome
func renderBeatIndicator(beat int) string {
	indicators := []string{"○", "○", "○", "○"}
	indicators[beat-1] = "●" // Highlight current beat

	// First beat is emphasized
	if beat == 1 {
		indicators[0] = "◉"
	}

	return strings.Join(indicators, " ")
}

// renderProgressBar creates a progress bar
func renderProgressBar(current, total int) string {
	barWidth := 20
	filled := int(float64(current) / float64(total) * float64(barWidth))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	return bar
}
