package player

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"backing-tracks/midi"
	"backing-tracks/parser"
)

// RealtimePlayer handles real-time MIDI playback with FluidSynth
type RealtimePlayer struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	playbackData *midi.PlaybackData
	track        *parser.Track

	// Playback state
	mu              sync.Mutex
	playing         bool
	paused          bool
	startTime       time.Time
	pausedAt        time.Time
	pausedTotal     time.Duration
	seekOffset      time.Duration
	lastEventIdx    int
	activeNotes     map[noteKey]bool // Track active notes for cleanup
	transposeOffset int              // Semitones to transpose
	capoPosition    int              // Capo fret position (0 = no capo)
	mutedTracks     [4]bool          // 0=drums, 1=bass, 2=chords, 3=melody

	// Loop state
	loopEnabled  bool // Whether loop is active
	loopStartBar int  // First bar of loop (inclusive)
	loopEndBar   int  // Last bar of loop (exclusive)
	loopLength   int  // Number of bars in loop (1-9)

	// Speed state
	tempoOffset int // BPM offset from original tempo (e.g., +10 or -20)

	// Control channels
	stopChan chan struct{}
	stopOnce sync.Once
}

type noteKey struct {
	channel uint8
	note    uint8
}

// GMInstruments maps friendly instrument names to General MIDI program numbers
var GMInstruments = map[string]int{
	// Pianos
	"piano":          0,
	"acoustic_piano": 0,
	"bright_piano":   1,
	"electric_piano": 4,
	"honky_tonk":     3,
	"harpsichord":    6,
	"clavinet":       7,

	// Guitars
	"nylon_guitar":    24,
	"steel_guitar":    25,
	"jazz_guitar":     26,
	"clean_guitar":    27,
	"muted_guitar":    28,
	"overdrive":       29,
	"distortion":      30,
	"harmonics":       31,

	// Bass
	"acoustic_bass":  32,
	"fingered_bass":  33,
	"picked_bass":    34,
	"fretless_bass":  35,
	"slap_bass":      36,
	"synth_bass":     38,

	// Strings
	"violin":         40,
	"viola":          41,
	"cello":          42,
	"contrabass":     43,
	"strings":        48,
	"slow_strings":   49,

	// Brass
	"trumpet":        56,
	"trombone":       57,
	"tuba":           58,
	"french_horn":    60,
	"brass":          61,
	"synth_brass":    62,

	// Woodwinds
	"soprano_sax":    64,
	"alto_sax":       65,
	"tenor_sax":      66,
	"baritone_sax":   67,
	"oboe":           68,
	"clarinet":       71,
	"flute":          73,
	"pan_flute":      75,

	// Synth
	"synth_lead":     80,
	"synth_pad":      88,

	// Organ
	"organ":          16,
	"church_organ":   19,
	"reed_organ":     20,
	"accordion":      21,
	"harmonica":      22,
	"bandoneon":      23,
}

// getGMProgram returns the GM program number for an instrument name
func getGMProgram(name string, defaultProg int) int {
	if name == "" {
		return defaultProg
	}
	if prog, ok := GMInstruments[name]; ok {
		return prog
	}
	return defaultProg
}

// NewRealtimePlayer creates a new real-time player
func NewRealtimePlayer(track *parser.Track, soundFont string) (*RealtimePlayer, error) {
	// Generate playback data
	playbackData := midi.GeneratePlaybackData(track)

	// Start FluidSynth in interactive mode
	cmd := exec.Command("fluidsynth",
		"-a", "pulseaudio", // or "alsa"
		"-q",               // Quiet mode
		"-s",               // Start as server (interactive)
		"-g", "1.0",        // Gain
		soundFont,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Discard stdout/stderr
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fluidsynth: %w", err)
	}

	// Give FluidSynth a moment to initialize
	time.Sleep(200 * time.Millisecond)

	// Set up instruments
	player := &RealtimePlayer{
		cmd:          cmd,
		stdin:        stdin,
		playbackData: playbackData,
		track:        track,
		activeNotes:  make(map[noteKey]bool),
		capoPosition: track.Info.Capo, // Initialize from track
		stopChan:     make(chan struct{}),
	}

	// Set program changes for each channel based on track settings
	chordsInstrument := ""
	if track.Rhythm != nil {
		chordsInstrument = track.Rhythm.Instrument
	}
	bassInstrument := ""
	if track.Bass != nil {
		bassInstrument = track.Bass.Instrument
	}
	melodyInstrument := ""
	if track.Melody != nil {
		melodyInstrument = track.Melody.Instrument
	}

	player.sendCommand(fmt.Sprintf("prog 0 %d", getGMProgram(chordsInstrument, 0)))  // Chords (default: piano)
	player.sendCommand(fmt.Sprintf("prog 1 %d", getGMProgram(bassInstrument, 33)))   // Bass (default: fingered bass)
	player.sendCommand(fmt.Sprintf("prog 2 %d", getGMProgram(melodyInstrument, 25))) // Melody (default: steel guitar)

	return player, nil
}

// sendCommand sends a command to FluidSynth
func (p *RealtimePlayer) sendCommand(cmd string) error {
	_, err := fmt.Fprintf(p.stdin, "%s\n", cmd)
	return err
}

// Start begins playback
func (p *RealtimePlayer) Start() {
	p.mu.Lock()
	p.playing = true
	p.paused = false
	p.startTime = time.Now()
	p.pausedTotal = 0
	p.seekOffset = 0
	p.lastEventIdx = 0
	p.mu.Unlock()

	go p.playbackLoop()
}

// playbackLoop is the main playback goroutine
func (p *RealtimePlayer) playbackLoop() {
	ticker := time.NewTicker(5 * time.Millisecond) // Check every 5ms for precise timing
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			p.allNotesOff()
			return
		case <-ticker.C:
			p.mu.Lock()
			if !p.playing || p.paused {
				p.mu.Unlock()
				continue
			}

			// Calculate current tick position (speed-adjusted)
			elapsed := p.getSpeedAdjustedElapsed()
			currentTick := p.playbackData.TimeToTick(elapsed)

			// Check for loop: if enabled and we've passed the loop end, jump back to loop start
			if p.loopEnabled && p.loopEndBar > 0 {
				loopEndTick := p.playbackData.BarToTick(p.loopEndBar)
				if currentTick >= loopEndTick {
					// Jump back to loop start
					p.seekToBarInternal(p.loopStartBar)
					p.mu.Unlock()
					continue
				}
			}

			// Check if we've reached the end
			if currentTick >= p.playbackData.TotalTicks {
				p.mu.Unlock()
				p.allNotesOff()
				return
			}

			// Play events up to current tick
			for p.lastEventIdx < len(p.playbackData.Events) {
				evt := p.playbackData.Events[p.lastEventIdx]
				if evt.Tick > currentTick {
					break
				}
				p.playEvent(evt)
				p.lastEventIdx++
			}

			p.mu.Unlock()
		}
	}
}

// playEvent sends a single event to FluidSynth
func (p *RealtimePlayer) playEvent(evt midi.PlaybackEvent) {
	// Check if track is muted
	// Channel mapping: 9=drums(0), 1=bass(1), 0=chords(2), 2=melody(3)
	trackIdx := -1
	switch evt.Channel {
	case 9:
		trackIdx = 0 // drums
	case 1:
		trackIdx = 1 // bass
	case 0:
		trackIdx = 2 // chords
	case 2:
		trackIdx = 3 // melody
	}
	if trackIdx >= 0 && p.mutedTracks[trackIdx] {
		return // Skip muted track
	}

	// Apply capo and transpose (except for drums on channel 9)
	note := evt.Note
	if evt.Channel != 9 {
		// Capo shifts pitch up, transpose can shift either direction
		offset := p.capoPosition + p.transposeOffset
		if offset != 0 {
			transposed := int(note) + offset
			if transposed < 0 {
				transposed = 0
			} else if transposed > 127 {
				transposed = 127
			}
			note = uint8(transposed)
		}
	}

	key := noteKey{evt.Channel, note}
	if evt.IsNoteOn {
		p.sendCommand(fmt.Sprintf("noteon %d %d %d", evt.Channel, note, evt.Velocity))
		p.activeNotes[key] = true
	} else {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", evt.Channel, note))
		delete(p.activeNotes, key)
	}
}

// Pause pauses playback
func (p *RealtimePlayer) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.paused {
		p.paused = true
		p.pausedAt = time.Now()
		// Silence all notes
		for key := range p.activeNotes {
			p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
		}
	}
}

// Resume resumes playback
func (p *RealtimePlayer) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.paused {
		p.pausedTotal += time.Since(p.pausedAt)
		p.paused = false
	}
}

// TogglePause toggles pause state
func (p *RealtimePlayer) TogglePause() {
	p.mu.Lock()
	paused := p.paused
	p.mu.Unlock()

	if paused {
		p.Resume()
	} else {
		p.Pause()
	}
}

// SeekToBar seeks to a specific bar
func (p *RealtimePlayer) SeekToBar(bar int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if bar < 0 {
		bar = 0
	}
	if bar >= p.playbackData.TotalBars {
		bar = p.playbackData.TotalBars - 1
	}

	// Stop all current notes
	for key := range p.activeNotes {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
	}
	p.activeNotes = make(map[noteKey]bool)

	// Calculate target tick
	targetTick := p.playbackData.BarToTick(bar)
	targetTime := p.playbackData.TickToTime(targetTick)

	// Adjust seek offset to jump to target
	p.seekOffset = targetTime - (time.Since(p.startTime) - p.pausedTotal)

	// Find the event index for the new position
	p.lastEventIdx = 0
	for i, evt := range p.playbackData.Events {
		if evt.Tick >= targetTick {
			p.lastEventIdx = i
			break
		}
	}
}

// SeekRelative seeks by a number of bars (positive = forward, negative = backward)
func (p *RealtimePlayer) SeekRelative(bars int) {
	p.mu.Lock()
	currentBar := p.getCurrentBar()
	p.mu.Unlock()

	p.SeekToBar(currentBar + bars)
}

// seekToBarInternal seeks to a bar (must be called with lock held)
func (p *RealtimePlayer) seekToBarInternal(bar int) {
	if bar < 0 {
		bar = 0
	}
	if bar >= p.playbackData.TotalBars {
		bar = p.playbackData.TotalBars - 1
	}

	// Stop all current notes
	for key := range p.activeNotes {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
	}
	p.activeNotes = make(map[noteKey]bool)

	// Calculate target tick
	targetTick := p.playbackData.BarToTick(bar)
	targetTime := p.playbackData.TickToTime(targetTick)

	// Adjust seek offset to jump to target
	p.seekOffset = targetTime - (time.Since(p.startTime) - p.pausedTotal)

	// Find the event index for the new position
	p.lastEventIdx = 0
	for i, evt := range p.playbackData.Events {
		if evt.Tick >= targetTick {
			p.lastEventIdx = i
			break
		}
	}
}

// SetLoop sets or clears the loop. length=0 disables looping.
// If length > 0, sets a loop from current bar for 'length' bars.
func (p *RealtimePlayer) SetLoop(length int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if length <= 0 {
		// Disable loop
		p.loopEnabled = false
		p.loopStartBar = 0
		p.loopEndBar = 0
		p.loopLength = 0
		return
	}

	// Enable loop from current bar
	currentBar := p.getCurrentBar()
	p.loopStartBar = currentBar
	p.loopEndBar = currentBar + length
	if p.loopEndBar > p.playbackData.TotalBars {
		p.loopEndBar = p.playbackData.TotalBars
	}
	p.loopLength = length
	p.loopEnabled = true
}

// ToggleLoop toggles loop of specified length. If already looping with same length, disables.
func (p *RealtimePlayer) ToggleLoop(length int) {
	p.mu.Lock()
	currentLength := p.loopLength
	enabled := p.loopEnabled
	p.mu.Unlock()

	if enabled && currentLength == length {
		// Same length - toggle off
		p.SetLoop(0)
	} else {
		// Different length or not enabled - set new loop
		p.SetLoop(length)
	}
}

// GetLoop returns the current loop state: enabled, startBar, endBar, length
func (p *RealtimePlayer) GetLoop() (enabled bool, startBar, endBar, length int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.loopEnabled, p.loopStartBar, p.loopEndBar, p.loopLength
}

// AdjustTempo adjusts the playback tempo by the given BPM delta (e.g., +5 or -5)
// Effective tempo is clamped to minimum 20 BPM
func (p *RealtimePlayer) AdjustTempo(deltaBPM int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	newOffset := p.tempoOffset + deltaBPM
	effectiveTempo := p.playbackData.Tempo + newOffset
	// Clamp to minimum 20 BPM
	if effectiveTempo < 20 {
		newOffset = 20 - p.playbackData.Tempo
	}
	p.tempoOffset = newOffset
}

// GetTempo returns the current effective tempo and the offset from original
func (p *RealtimePlayer) GetTempo() (effectiveBPM int, offset int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playbackData.Tempo + p.tempoOffset, p.tempoOffset
}

// GetCurrentSection returns the section at the current playback position
func (p *RealtimePlayer) GetCurrentSection() (name string, startBar, endBar int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	currentBar := p.getCurrentBar()
	section := p.playbackData.GetSectionAtBar(currentBar)
	if section == nil {
		return "", 0, 0
	}
	return section.Name, section.StartBar, section.EndBar
}

// LoopCurrentSection sets the loop to the current section
func (p *RealtimePlayer) LoopCurrentSection() {
	p.mu.Lock()
	defer p.mu.Unlock()

	currentBar := p.getCurrentBar()
	section := p.playbackData.GetSectionAtBar(currentBar)
	if section == nil {
		// No section at current position - disable loop
		p.loopEnabled = false
		p.loopStartBar = 0
		p.loopEndBar = 0
		p.loopLength = 0
		return
	}

	// If already looping this section, toggle off
	if p.loopEnabled && p.loopStartBar == section.StartBar && p.loopEndBar == section.EndBar {
		p.loopEnabled = false
		p.loopStartBar = 0
		p.loopEndBar = 0
		p.loopLength = 0
		return
	}

	// Set loop to current section
	p.loopEnabled = true
	p.loopStartBar = section.StartBar
	p.loopEndBar = section.EndBar
	p.loopLength = section.EndBar - section.StartBar
}

// GetSections returns all sections in the track
func (p *RealtimePlayer) GetSections() []struct{ Name string; StartBar, EndBar int } {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]struct{ Name string; StartBar, EndBar int }, len(p.playbackData.Sections))
	for i, s := range p.playbackData.Sections {
		result[i] = struct{ Name string; StartBar, EndBar int }{s.Name, s.StartBar, s.EndBar}
	}
	return result
}

// getSpeedAdjustedElapsed returns the elapsed playback time adjusted for tempo changes (must be called with lock held)
func (p *RealtimePlayer) getSpeedAdjustedElapsed() time.Duration {
	realElapsed := time.Since(p.startTime) - p.pausedTotal + p.seekOffset
	if realElapsed < 0 {
		realElapsed = 0
	}
	// Calculate speed multiplier from tempo offset
	// e.g., original 120 BPM + 10 offset = 130 BPM effective = 130/120 = 1.083x speed
	effectiveTempo := float64(p.playbackData.Tempo + p.tempoOffset)
	originalTempo := float64(p.playbackData.Tempo)
	speedMultiplier := effectiveTempo / originalTempo
	return time.Duration(float64(realElapsed) * speedMultiplier)
}

// getCurrentBar returns the current bar (must be called with lock held)
func (p *RealtimePlayer) getCurrentBar() int {
	elapsed := p.getSpeedAdjustedElapsed()
	currentTick := p.playbackData.TimeToTick(elapsed)
	return int(currentTick / p.playbackData.TicksPerBar)
}

// GetCurrentBar returns the current bar (thread-safe)
func (p *RealtimePlayer) GetCurrentBar() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.getCurrentBar()
}

// IsPaused returns whether playback is paused
func (p *RealtimePlayer) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}

// Transpose adjusts the transpose offset by the given semitones
func (p *RealtimePlayer) Transpose(semitones int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop all current notes before changing transpose
	for key := range p.activeNotes {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
	}
	p.activeNotes = make(map[noteKey]bool)

	p.transposeOffset += semitones
}

// GetTranspose returns the current transpose offset in semitones
func (p *RealtimePlayer) GetTranspose() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transposeOffset
}

// SetCapo sets the capo position (0 = no capo)
func (p *RealtimePlayer) SetCapo(fret int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if fret < 0 {
		fret = 0
	} else if fret > 12 {
		fret = 12
	}

	// Stop all current notes before changing capo
	for key := range p.activeNotes {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
	}
	p.activeNotes = make(map[noteKey]bool)

	p.capoPosition = fret
}

// GetCapo returns the current capo position
func (p *RealtimePlayer) GetCapo() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.capoPosition
}

// ToggleTrackMute toggles mute state for a track (0=drums, 1=bass, 2=chords, 3=melody)
func (p *RealtimePlayer) ToggleTrackMute(track int) {
	if track < 0 || track > 3 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.mutedTracks[track] = !p.mutedTracks[track]

	// If muting, stop all notes on that channel
	if p.mutedTracks[track] {
		// Map track to channel
		var channel uint8
		switch track {
		case 0:
			channel = 9 // drums
		case 1:
			channel = 1 // bass
		case 2:
			channel = 0 // chords
		case 3:
			channel = 2 // melody
		}
		// Stop notes on this channel
		for key := range p.activeNotes {
			if key.channel == channel {
				p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
				delete(p.activeNotes, key)
			}
		}
	}
}

// IsTrackMuted returns whether a track is muted (0=drums, 1=bass, 2=chords, 3=melody)
func (p *RealtimePlayer) IsTrackMuted(track int) bool {
	if track < 0 || track > 3 {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mutedTracks[track]
}

// allNotesOff sends note-off for all channels
func (p *RealtimePlayer) allNotesOff() {
	// Turn off any active notes
	for key := range p.activeNotes {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", key.channel, key.note))
	}
	p.activeNotes = make(map[noteKey]bool)

	// Also send all-notes-off for safety
	for ch := 0; ch < 16; ch++ {
		p.sendCommand(fmt.Sprintf("cc %d 123 0", ch)) // All notes off
	}
}

// Stop stops playback and cleans up
func (p *RealtimePlayer) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})

	p.allNotesOff()
	p.sendCommand("quit")
	p.stdin.Close()

	// Wait for FluidSynth with timeout
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case <-done:
		// FluidSynth exited normally
	case <-time.After(2 * time.Second):
		// Timeout - force kill
		p.cmd.Process.Kill()
		<-done
	}
}

// GetPlaybackState returns current playback state for TUI sync
func (p *RealtimePlayer) GetPlaybackState() (bar int, beat int, strum int, paused bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate elapsed time (speed-adjusted)
	var elapsed time.Duration
	if p.paused {
		// When paused, use time up to when pause happened
		realElapsed := p.pausedAt.Sub(p.startTime) - p.pausedTotal + p.seekOffset
		if realElapsed < 0 {
			realElapsed = 0
		}
		effectiveTempo := float64(p.playbackData.Tempo + p.tempoOffset)
		originalTempo := float64(p.playbackData.Tempo)
		speedMultiplier := effectiveTempo / originalTempo
		elapsed = time.Duration(float64(realElapsed) * speedMultiplier)
	} else {
		elapsed = p.getSpeedAdjustedElapsed()
	}

	currentTick := p.playbackData.TimeToTick(elapsed)
	ticksPerBeat := p.playbackData.TicksPerBar / 4

	bar = int(currentTick / p.playbackData.TicksPerBar)
	beat = int((currentTick % p.playbackData.TicksPerBar) / ticksPerBeat)

	// Calculate strum position based on rhythm style
	strumsPerBar := 8 // Default for 8th notes
	if p.track != nil && p.track.Rhythm != nil {
		switch p.track.Rhythm.Style {
		case "sixteenth", "funk_16th", "funk_muted", "dust_in_wind", "landslide", "pima", "pima_reverse":
			strumsPerBar = 16
		}
	}
	ticksPerStrum := p.playbackData.TicksPerBar / uint32(strumsPerBar)
	strum = int((currentTick % p.playbackData.TicksPerBar) / ticksPerStrum)

	paused = p.paused

	return
}

// WaitForInput waits for user input to control playback (for non-TUI mode)
func (p *RealtimePlayer) WaitForInput() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Controls: [space] pause/resume, [n] next bar, [p] prev bar, [q] quit")
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "", " ":
			p.TogglePause()
		case "n":
			p.SeekRelative(1)
		case "p":
			p.SeekRelative(-1)
		case "q":
			return
		}
	}
}
