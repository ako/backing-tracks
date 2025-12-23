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
	mu            sync.Mutex
	playing       bool
	paused        bool
	startTime     time.Time
	pausedAt      time.Time
	pausedTotal   time.Duration
	seekOffset    time.Duration
	lastEventIdx  int
	activeNotes   map[noteKey]bool // Track active notes for cleanup

	// Control channels
	stopChan chan struct{}
}

type noteKey struct {
	channel uint8
	note    uint8
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
		stopChan:     make(chan struct{}),
	}

	// Set program changes for each channel
	player.sendCommand("prog 0 0")   // Piano for chords
	player.sendCommand("prog 1 33")  // Fingered bass
	player.sendCommand("prog 2 25")  // Steel guitar for melody

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

			// Calculate current tick position
			elapsed := time.Since(p.startTime) - p.pausedTotal + p.seekOffset
			if elapsed < 0 {
				elapsed = 0
			}
			currentTick := p.playbackData.TimeToTick(elapsed)

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
	key := noteKey{evt.Channel, evt.Note}
	if evt.IsNoteOn {
		p.sendCommand(fmt.Sprintf("noteon %d %d %d", evt.Channel, evt.Note, evt.Velocity))
		p.activeNotes[key] = true
	} else {
		p.sendCommand(fmt.Sprintf("noteoff %d %d", evt.Channel, evt.Note))
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

// getCurrentBar returns the current bar (must be called with lock held)
func (p *RealtimePlayer) getCurrentBar() int {
	elapsed := time.Since(p.startTime) - p.pausedTotal + p.seekOffset
	if elapsed < 0 {
		elapsed = 0
	}
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
	close(p.stopChan)
	p.allNotesOff()
	p.sendCommand("quit")
	p.stdin.Close()
	p.cmd.Wait()
}

// GetPlaybackState returns current playback state for TUI sync
func (p *RealtimePlayer) GetPlaybackState() (bar int, beat int, strum int, paused bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.startTime) - p.pausedTotal + p.seekOffset
	if p.paused {
		elapsed = p.pausedAt.Sub(p.startTime) - p.pausedTotal + p.seekOffset
	}
	if elapsed < 0 {
		elapsed = 0
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
