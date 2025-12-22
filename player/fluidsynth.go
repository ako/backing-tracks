package player

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"backing-tracks/display"
	"backing-tracks/parser"
)

// PlayMIDIWithDisplay plays a MIDI file using FluidSynth with live visual display
func PlayMIDIWithDisplay(midiFile string, track *parser.Track) error {
	// Check if FluidSynth is installed
	if _, err := exec.LookPath("fluidsynth"); err != nil {
		return fmt.Errorf("fluidsynth not found: please install with 'sudo apt install fluidsynth'")
	}

	// Find a SoundFont file
	soundFont, err := findSoundFont()
	if err != nil {
		return err
	}

	fmt.Printf("Using SoundFont: %s\n", soundFont)
	fmt.Println()

	// Create and start live display
	liveDisplay := display.NewLiveDisplay(track)
	liveDisplay.Start()
	defer liveDisplay.Stop()

	// Build FluidSynth command
	// -ni: no interactive mode
	// -r 48000: sample rate
	// -g 1.0: gain
	// -q: quiet mode (suppress FluidSynth output)
	cmd := exec.Command("fluidsynth",
		"-ni",           // No interactive mode
		"-q",            // Quiet mode
		"-r", "48000",   // Sample rate
		"-g", "1.0",     // Gain
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
	soundFont, err := findSoundFont()
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

// findSoundFont locates a SoundFont file on the system
func findSoundFont() (string, error) {
	// Common SoundFont locations on Linux
	locations := []string{
		"/usr/share/sounds/sf2/FluidR3_GM.sf2",
		"/usr/share/sounds/sf2/default.sf2",
		"/usr/share/soundfonts/FluidR3_GM.sf2",
		"/usr/share/soundfonts/default.sf2",
		"/usr/share/soundfonts/default-GM.sf2",
		"/usr/share/sounds/sf2/TimGM6mb.sf2",
		// Add more locations as needed
	}

	for _, loc := range locations {
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
		"  sudo apt install fluid-soundfont-gm")
}
