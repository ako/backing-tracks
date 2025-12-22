package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"backing-tracks/display"
	"backing-tracks/midi"
	"backing-tracks/parser"
	"backing-tracks/player"
	"backing-tracks/strudel"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	filename := os.Args[2]

	switch command {
	case "play":
		playTrack(filename)
	case "export":
		outputPath := ""
		if len(os.Args) >= 4 {
			outputPath = os.Args[3]
		}
		exportTrack(filename, outputPath)
	case "strudel":
		outputPath := ""
		if len(os.Args) >= 4 {
			outputPath = os.Args[3]
		}
		exportStrudel(filename, outputPath)
	default:
		printUsage()
		os.Exit(1)
	}
}

func playTrack(filename string) {
	// Parse BTML file
	track, err := parser.LoadTrack(filename)
	if err != nil {
		fmt.Printf("Error loading track: %v\n", err)
		os.Exit(1)
	}

	// Display track info in terminal
	display.ShowTrack(track)

	// Generate MIDI file from track
	midiFile, err := midi.GenerateFromTrack(track)
	if err != nil {
		fmt.Printf("Error generating MIDI: %v\n", err)
		os.Exit(1)
	}

	// Play via FluidSynth with live display
	fmt.Println("♪ Playing... (Press Ctrl+C to stop)\n")
	if err := player.PlayMIDIWithDisplay(midiFile, track); err != nil {
		fmt.Printf("Error playing: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n\n✓ Playback complete!")
}

func exportTrack(filename, outputPath string) {
	// Parse BTML file
	track, err := parser.LoadTrack(filename)
	if err != nil {
		fmt.Printf("Error loading track: %v\n", err)
		os.Exit(1)
	}

	// Display track info
	display.ShowTrack(track)

	// Generate MIDI file
	tmpFile, err := midi.GenerateFromTrack(track)
	if err != nil {
		fmt.Printf("Error generating MIDI: %v\n", err)
		os.Exit(1)
	}

	// Determine output path
	if outputPath == "" {
		// Default: same name as input with .mid extension
		base := filepath.Base(filename)
		ext := filepath.Ext(base)
		outputPath = strings.TrimSuffix(base, ext) + ".mid"
	}

	// Copy from temp to output
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		fmt.Printf("Error reading MIDI: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Printf("Error writing MIDI: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Exported to: %s\n", outputPath)
}

func exportStrudel(filename, outputPath string) {
	// Parse BTML file
	track, err := parser.LoadTrack(filename)
	if err != nil {
		fmt.Printf("Error loading track: %v\n", err)
		os.Exit(1)
	}

	// Display track info
	display.ShowTrack(track)

	// Generate Strudel code
	code := strudel.GenerateStrudel(track)

	// Determine output path
	if outputPath == "" {
		// Default: same name as input with .js extension
		base := filepath.Base(filename)
		ext := filepath.Ext(base)
		outputPath = strings.TrimSuffix(base, ext) + ".strudel.js"
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		fmt.Printf("Error writing Strudel file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Exported to: %s\n", outputPath)
	fmt.Println("\nPaste the code into https://strudel.cc to play!")
}

func printUsage() {
	fmt.Println("Backing Tracks Player v0.4")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  backing-tracks play <file.btml>            Play backing track")
	fmt.Println("  backing-tracks export <file.btml> [out]    Export to MIDI file")
	fmt.Println("  backing-tracks strudel <file.btml> [out]   Export to Strudel code")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  backing-tracks play examples/blues-full.btml")
	fmt.Println("  backing-tracks export examples/blues-full.btml")
	fmt.Println("  backing-tracks export examples/blues-full.btml my-track.mid")
	fmt.Println("  backing-tracks strudel examples/blues-full.btml")
}
