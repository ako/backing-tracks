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

// Global soundfont path (can be set via --soundfont flag)
var soundFontPath string

func main() {
	args := parseArgs(os.Args[1:])

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "play":
		if len(args) < 2 {
			fmt.Println("Error: play requires a BTML file")
			printUsage()
			os.Exit(1)
		}
		playTrack(args[1])
	case "export":
		if len(args) < 2 {
			fmt.Println("Error: export requires a BTML file")
			printUsage()
			os.Exit(1)
		}
		outputPath := ""
		if len(args) >= 3 {
			outputPath = args[2]
		}
		exportTrack(args[1], outputPath)
	case "strudel":
		if len(args) < 2 {
			fmt.Println("Error: strudel requires a BTML file")
			printUsage()
			os.Exit(1)
		}
		outputPath := ""
		if len(args) >= 3 {
			outputPath = args[2]
		}
		exportStrudel(args[1], outputPath)
	case "soundfonts":
		listSoundFonts()
	default:
		printUsage()
		os.Exit(1)
	}
}

// parseArgs extracts flags and returns remaining args
func parseArgs(args []string) []string {
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--soundfont" || arg == "-sf" {
			if i+1 < len(args) {
				soundFontPath = args[i+1]
				i++ // Skip next arg
			} else {
				fmt.Println("Error: --soundfont requires a path")
				os.Exit(1)
			}
		} else if strings.HasPrefix(arg, "--soundfont=") {
			soundFontPath = strings.TrimPrefix(arg, "--soundfont=")
		} else if strings.HasPrefix(arg, "-sf=") {
			soundFontPath = strings.TrimPrefix(arg, "-sf=")
		} else if arg == "--help" || arg == "-h" {
			printUsage()
			os.Exit(0)
		} else {
			remaining = append(remaining, arg)
		}
	}

	// Also check environment variable
	if soundFontPath == "" {
		soundFontPath = os.Getenv("SOUNDFONT")
	}

	return remaining
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
	if err := player.PlayMIDIWithDisplay(midiFile, track, soundFontPath); err != nil {
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

func listSoundFonts() {
	fmt.Println("Available SoundFonts:")
	fmt.Println()

	found := player.ListSoundFonts()

	if len(found) == 0 {
		fmt.Println("  No SoundFonts found!")
		fmt.Println()
		fmt.Println("Install the default SoundFont:")
		fmt.Println("  sudo apt install fluid-soundfont-gm")
		fmt.Println()
		fmt.Println("Or download better SoundFonts:")
		fmt.Println("  - FluidR3 GM (140MB): https://member.keymusician.com/Member/FluidR3_GM/")
		fmt.Println("  - SGM-V2.01 (235MB):  https://musical-artifacts.com/artifacts/855")
		fmt.Println("  - Timbres of Heaven: https://midkar.com/soundfonts/")
		fmt.Println()
		fmt.Println("Place .sf2 files in ./soundfonts/ or specify with --soundfont flag")
	} else {
		for _, sf := range found {
			fmt.Printf("  %s\n", sf)
		}
		fmt.Println()
		fmt.Println("Use with: ./backing-tracks play --soundfont <path> <file.btml>")
	}
}

func printUsage() {
	fmt.Println("Backing Tracks Player v0.5")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  backing-tracks play <file.btml>              Play backing track")
	fmt.Println("  backing-tracks export <file.btml> [out]      Export to MIDI file")
	fmt.Println("  backing-tracks strudel <file.btml> [out]     Export to Strudel code")
	fmt.Println("  backing-tracks soundfonts                    List available SoundFonts")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --soundfont, -sf <path>   Use custom SoundFont (.sf2 file)")
	fmt.Println("  --help, -h                Show this help")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SOUNDFONT                 Default SoundFont path")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  backing-tracks play examples/blues-full.btml")
	fmt.Println("  backing-tracks play --soundfont ~/soundfonts/SGM.sf2 examples/edm-808.btml")
	fmt.Println("  backing-tracks export examples/blues-full.btml my-track.mid")
	fmt.Println("  backing-tracks strudel examples/blues-full.btml")
	fmt.Println()
	fmt.Println("SoundFont tips:")
	fmt.Println("  Place .sf2 files in ./soundfonts/ directory for auto-detection")
	fmt.Println("  Set SOUNDFONT env var for a permanent default")
}
