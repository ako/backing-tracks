package player

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"backing-tracks/display"
	"backing-tracks/parser"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// PlayMIDIWithDisplay plays a MIDI file using FluidSynth with live TUI display
func PlayMIDIWithDisplay(midiFile string, track *parser.Track, customSoundFont string) error {
	// Check if FluidSynth is installed
	if _, err := exec.LookPath("fluidsynth"); err != nil {
		return fmt.Errorf("fluidsynth not found: please install with 'sudo apt install fluidsynth'")
	}

	// Find a SoundFont file
	soundFont, err := findSoundFont(customSoundFont)
	if err != nil {
		return err
	}

	fmt.Printf("Using SoundFont: %s\n", soundFont)
	fmt.Println()

	// Check if we have a TTY - if not, use legacy display
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return playWithLegacyDisplay(midiFile, track, soundFont)
	}

	// Create real-time player
	player, err := NewRealtimePlayer(track, soundFont)
	if err != nil {
		// Fall back to file-based playback if real-time fails
		fmt.Println("Real-time playback unavailable, using file-based playback...")
		return playWithFileBasedTUI(midiFile, track, soundFont)
	}
	defer player.Stop()

	// Create TUI model and connect to player
	tuiModel := display.NewTUIModel(track)
	tuiModel.SetPlayer(player)

	// Start playback
	player.Start()

	// Run the TUI
	p := tea.NewProgram(tuiModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// playWithFileBasedTUI is the fallback when real-time playback isn't available
func playWithFileBasedTUI(midiFile string, track *parser.Track, soundFont string) error {
	// Create TUI model
	tuiModel := display.NewTUIModel(track)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to signal when FluidSynth finishes
	done := make(chan error, 1)

	// Build FluidSynth command with context
	cmd := exec.CommandContext(ctx, "fluidsynth",
		"-ni",         // No interactive mode
		"-q",          // Quiet mode
		"-r", "48000", // Sample rate
		"-g", "1.0",   // Gain
		soundFont,
		midiFile,
	)

	// Discard stdout/stderr to keep display clean
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	// Start FluidSynth in background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start fluidsynth: %w", err)
	}

	// Wait for FluidSynth in goroutine
	go func() {
		done <- cmd.Wait()
	}()

	// Run the TUI
	p := tea.NewProgram(tuiModel, tea.WithAltScreen())

	// Run TUI in goroutine
	tuiDone := make(chan error, 1)
	go func() {
		_, err := p.Run()
		tuiDone <- err
	}()

	// Wait for either FluidSynth to finish or TUI to quit
	select {
	case err := <-done:
		// FluidSynth finished - stop TUI
		p.Send(tea.Quit())
		<-tuiDone
		if err != nil && !tuiModel.IsQuitting() {
			return fmt.Errorf("fluidsynth error: %w", err)
		}
	case err := <-tuiDone:
		// TUI quit (user pressed q) - stop FluidSynth
		cancel()
		<-done
		if err != nil {
			return err
		}
	}

	return nil
}

// playWithLegacyDisplay uses the old ANSI-based display (for non-TTY environments)
func playWithLegacyDisplay(midiFile string, track *parser.Track, soundFont string) error {
	// Create and start legacy live display
	liveDisplay := display.NewLiveDisplay(track)
	liveDisplay.Start()
	defer liveDisplay.Stop()

	// Build FluidSynth command
	cmd := exec.Command("fluidsynth",
		"-ni",         // No interactive mode
		"-q",          // Quiet mode
		"-r", "48000", // Sample rate
		"-g", "1.0",   // Gain
		soundFont,
		midiFile,
	)

	// Discard stdout/stderr to keep display clean
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	// Run and wait for completion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fluidsynth error: %w", err)
	}

	return nil
}

// PlayMIDI plays a MIDI file using FluidSynth (legacy without display)
func PlayMIDI(midiFile string) error {
	// Check if FluidSynth is installed
	if _, err := exec.LookPath("fluidsynth"); err != nil {
		return fmt.Errorf("fluidsynth not found: please install with 'sudo apt install fluidsynth'")
	}

	// Find a SoundFont file
	soundFont, err := findSoundFont("")
	if err != nil {
		return err
	}

	fmt.Printf("Using SoundFont: %s\n", soundFont)

	// Build FluidSynth command
	// -ni: no interactive mode
	// -r 48000: sample rate
	// -g 1.0: gain
	cmd := exec.Command("fluidsynth",
		"-ni",           // No interactive mode
		"-r", "48000",   // Sample rate
		"-g", "1.0",     // Gain
		soundFont,
		midiFile,
	)

	// Connect stdout/stderr to see output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run and wait for completion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fluidsynth error: %w", err)
	}

	return nil
}

// ListSoundFonts returns all available soundfonts on the system
func ListSoundFonts() []string {
	var found []string

	// Check local soundfonts directory first
	localPatterns := []string{
		"./soundfonts/*.sf2",
		"./soundfonts/*.SF2",
	}

	for _, pattern := range localPatterns {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			found = append(found, matches...)
		}
	}

	// Check system locations
	systemLocations := []string{
		"/usr/share/sounds/sf2/FluidR3_GM.sf2",
		"/usr/share/sounds/sf2/default.sf2",
		"/usr/share/soundfonts/FluidR3_GM.sf2",
		"/usr/share/soundfonts/default.sf2",
		"/usr/share/soundfonts/default-GM.sf2",
		"/usr/share/sounds/sf2/TimGM6mb.sf2",
	}

	for _, loc := range systemLocations {
		if _, err := os.Stat(loc); err == nil {
			found = append(found, loc)
		}
	}

	// Find any other .sf2 files in system directories
	systemPatterns := []string{
		"/usr/share/sounds/sf2/*.sf2",
		"/usr/share/soundfonts/*.sf2",
		"~/.local/share/soundfonts/*.sf2",
	}

	for _, pattern := range systemPatterns {
		// Expand ~ to home directory
		if pattern[0] == '~' {
			home, _ := os.UserHomeDir()
			pattern = home + pattern[1:]
		}
		matches, err := filepath.Glob(pattern)
		if err == nil {
			for _, m := range matches {
				// Avoid duplicates
				isDup := false
				for _, f := range found {
					if f == m {
						isDup = true
						break
					}
				}
				if !isDup {
					found = append(found, m)
				}
			}
		}
	}

	return found
}

// findSoundFont locates a SoundFont file on the system
func findSoundFont(customPath string) (string, error) {
	// If custom path provided, use it
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			return customPath, nil
		}
		return "", fmt.Errorf("soundfont not found: %s", customPath)
	}

	// Check local soundfonts directory first (project-local)
	localPatterns := []string{
		"./soundfonts/*.sf2",
		"./soundfonts/*.SF2",
	}

	for _, pattern := range localPatterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0], nil
		}
	}

	// Check user's local soundfonts
	home, _ := os.UserHomeDir()
	userLocations := []string{
		filepath.Join(home, ".local/share/soundfonts"),
		filepath.Join(home, "soundfonts"),
	}

	for _, dir := range userLocations {
		pattern := filepath.Join(dir, "*.sf2")
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0], nil
		}
	}

	// Common SoundFont locations on Linux
	systemLocations := []string{
		"/usr/share/sounds/sf2/FluidR3_GM.sf2",
		"/usr/share/sounds/sf2/default.sf2",
		"/usr/share/soundfonts/FluidR3_GM.sf2",
		"/usr/share/soundfonts/default.sf2",
		"/usr/share/soundfonts/default-GM.sf2",
		"/usr/share/sounds/sf2/TimGM6mb.sf2",
	}

	for _, loc := range systemLocations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	// Try to find any .sf2 file
	patterns := []string{
		"/usr/share/sounds/sf2/*.sf2",
		"/usr/share/soundfonts/*.sf2",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0], nil
		}
	}

	return "", fmt.Errorf("no SoundFont (.sf2) file found. Please install fluid-soundfont-gm:\n" +
		"  sudo apt install fluid-soundfont-gm\n\n" +
		"Or place custom .sf2 files in ./soundfonts/ directory\n" +
		"Or specify with --soundfont flag")
}
